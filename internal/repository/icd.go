package repository

import (
	"backend/cmd/web/dto"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"sync"
	"time"
)

type ICDRepository interface {
	Find(input string) (*ICDMatches, error)
	List(size int) ([]ICDMatch, error)
}

type icdRepository struct {
	client       *http.Client
	clientID     string
	clientSecret string

	mu          sync.Mutex
	accessToken string
	expiry      time.Time
}

type ICDMatch struct {
	ID   string
	Name string
	Desc string
}

type ICDMatches struct {
	Matches []ICDMatch
}

func (i *icdRepository) fetchDescription(id string, ch chan string) {
	descriptionURL := "https://id.who.int/icd/release/11/2025-01/mms/" + id

	req, err := http.NewRequest("GET", descriptionURL, nil)
	if err != nil {
		log.Println("Error: " + err.Error())
		ch <- ""
		return
	}

	req.Header.Set("Authorization", "Bearer "+i.accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("API-Version", "v2")
	req.Header.Set("Accept-Language", "en")

	resp, err := i.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if err != nil {
			log.Println("Error: " + err.Error())
		} else {
			log.Println("Error: " + resp.Status)
		}
		ch <- ""
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error: " + err.Error())
		ch <- ""
		return
	}

	type definition struct {
		Value string `json:"@value"`
	}

	var descriptionResponse struct {
		Definition definition `json:"definition"`
	}

	if err := json.Unmarshal(body, &descriptionResponse); err != nil {
		log.Println("Error: " + err.Error())
		ch <- ""
		return
	}

	log.Println("Found description for id " + id + ": " + descriptionResponse.Definition.Value)
	ch <- descriptionResponse.Definition.Value
}

func (i *icdRepository) List(size int) ([]ICDMatch, error) {
	const searchURL = "https://id.who.int/icd/release/11/2025-01/mms/search"
	const numWorkers = 2 // Control the number of parallel requests

	if err := i.ensureToken(); err != nil {
		return nil, err
	}

	var allMatches []ICDMatch
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// Channels for managing the parallel search
	letters := make(chan string, len(alphabet))
	results := make(chan []ICDMatch, len(alphabet))
	errs := make(chan error, len(alphabet))

	// Start the worker goroutines
	for range numWorkers {
		go func() {
			for char := range letters {
				// Each worker handles one letter at a time
				matches, err := i.fetchMatchesForLetter(searchURL, char, size)
				if err != nil {
					errs <- err
					continue
				}
				results <- matches
			}
		}()
	}

	// Send letters to the workers
	for _, char := range alphabet {
		letters <- string(char)
	}
	close(letters) // No more letters will be sent

	// Collect results from the workers
	for range alphabet {
		select {
		case res := <-results:
			allMatches = append(allMatches, res...)
			// If we have enough results, we can stop
			if len(allMatches) >= size {
				return allMatches[:size], nil
			}
		case err := <-errs:
			// Handle the first error and return
			return nil, err
		}
	}

	return allMatches, nil
}

// fetchMatchesForLetter is a helper function to perform the actual API call for a single letter
func (i *icdRepository) fetchMatchesForLetter(searchURL, char string, size int) ([]ICDMatch, error) {
	needed := size // We don't know exactly how many we need, so we ask for the full size

	req, err := http.NewRequest("GET", searchURL+"?q="+url.QueryEscape(char)+"&maxList="+strconv.Itoa(needed)+"&flatResults=true&highlightingEnabled=false", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+i.accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("API-Version", "v2")
	req.Header.Set("Accept-Language", "en")

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status for search with '%s': %s", char, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed: %w", err)
	}

	var response dto.SearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	var matches []ICDMatch
	channels := make(map[int]chan string)
	for idx, entity := range response.DestinationEntities {

		var definition string
		for _, pv := range entity.MatchingPVs {
			if pv.PropertyID == "Definition" {
				definition = pv.Label
				break
			}
		}

		if definition == "" {
			ch := make(chan string)
			channels[idx] = ch
			parsedURL, err := url.Parse(entity.ID)
			if err != nil {
				log.Fatalln("Invalid id: %w", err)
			}
			id := path.Base(parsedURL.Path)

			log.Printf("Fetching description for code (%s) id: %s\n", entity.TheCode, entity.ID)
			go i.fetchDescription(id, ch)
		}
		matches = append(matches, ICDMatch{
			ID:   entity.TheCode,
			Name: entity.Title,
		})
	}

	for idx, ch := range channels {
		matches[idx].Desc = <-ch
	}

	return matches, nil
}

func (i *icdRepository) Find(input string) (*ICDMatches, error) {
	const searchURL = "https://id.who.int/icd/release/11/2025-01/mms/search"

	if err := i.ensureToken(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", searchURL+"?q="+url.QueryEscape(input)+"&subtreeFilterUsesFoundationDescendants=false&includeKeywordResult=false&useFlexisearch=false&flatResults=true&highlightingEnabled=false&medicalCodingMode=false&propertiesToBeSearched=Title%2CFullySpecifiedName%2CDefinition%2CIndexTerm", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+i.accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("API-Version", "v2")
	req.Header.Set("Accept-Language", "en")

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed: %w", err)
	}

	var response dto.SearchResponse

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	matches := make([]ICDMatch, 0, 5)
	channels := make(map[int]chan string)

	for idx, entity := range response.DestinationEntities {
		if idx == 5 {
			break
		}

		var definition string
		for _, pv := range entity.MatchingPVs {
			if pv.PropertyID == "Definition" {
				definition = pv.Label
				break
			}
		}

		if definition == "" {
			ch := make(chan string)
			channels[idx] = ch
			parsedURL, err := url.Parse(entity.ID)
			if err != nil {
				log.Fatalln("Invalid id: %w", err)
			}
			id := path.Base(parsedURL.Path)

			log.Printf("Fetching description for code (%s) id: %s\n", entity.TheCode, entity.ID)
			go i.fetchDescription(id, ch)
		}

		matches = append(matches, ICDMatch{
			ID:   entity.TheCode,
			Name: entity.Title,
		})
	}

	for idx, ch := range channels {
		matches[idx].Desc = <-ch
	}

	return &ICDMatches{
		Matches: matches,
	}, nil
}

func NewICDRepository(client *http.Client, clientID string, clientSecret string) ICDRepository {
	return &icdRepository{
		client:       client,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (i *icdRepository) getToken() error {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", i.clientID)
	data.Set("client_secret", i.clientSecret)
	data.Set("scope", "icdapi_access")

	const tokenURL = "https://icdaccessmanagement.who.int/connect/token"

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad token response: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body failed: %w", err)
	}

	var tokenResp dto.TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("decode token failed: %w", err)
	}

	i.accessToken = tokenResp.AccessToken
	i.expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)
	return nil
}

func (i *icdRepository) ensureToken() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if time.Now().After(i.expiry) || i.accessToken == "" {
		return i.getToken()
	}

	return nil
}
