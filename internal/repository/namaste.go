package repository

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
)

type namasteMatch struct {
	Type string
	ID   string
	Name string
	Desc string
}

type NamasteMatches struct {
	Diseases []namasteMatch
}

type NamasteRepository interface {
	CreateIndex(path string) error
	Find(input string) (*NamasteMatches, error)
}

type namasteRepository struct{}

func NewNamasteRepository() NamasteRepository {
	return &namasteRepository{}
}

// CreateIndex implements NamasteRepository.
func (n *namasteRepository) CreateIndex(path string) error {
	// Define 3 branches of traditional medicine
	branches := []string{"ayurveda", "unani", "siddha"}

	// Delete the old index
	os.RemoveAll(path)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(path, mapping)
	if err != nil {
		return fmt.Errorf("error creating bleve index: %w", err)
	}
	defer index.Close()

	log.Println("Starting to index documents...")
	for _, branch := range branches {
		// Get the traditional medicine csv from assets
		file, err := os.Open("assets/" + branch + ".csv")
		if err != nil {
			return fmt.Errorf("error opening asset: %w", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			return fmt.Errorf("error reading CSV records from %s: %w", branch, err)
		}

		for i, recordCSV := range records {
			if i == 0 { // Skip header row
				continue
			}

			var record struct {
				Type        string
				ID          string
				Code        string
				Term        string
				Diacritical string
				Native      string
				ShortDesc   string
				LongDesc    string
			}

			switch branch {
			case "ayurveda":
				record.Type = "ayurveda"
				record.ID = recordCSV[1]
				record.Code = recordCSV[2]
				record.Term = recordCSV[3]
				record.Diacritical = recordCSV[4]
				record.Native = recordCSV[5]
				record.ShortDesc = recordCSV[6]
				record.LongDesc = recordCSV[7]
			case "unani":
				record.Type = "unani"
				record.ID = recordCSV[1]
				record.Code = recordCSV[2]
				record.Native = recordCSV[3]
				record.Term = recordCSV[4]
				record.Diacritical = recordCSV[4]
				record.ShortDesc = recordCSV[5]
				record.LongDesc = recordCSV[6]
			case "siddha":
				record.Type = "siddha"
				record.ID = recordCSV[1]
				record.Code = recordCSV[2]
				record.Term = recordCSV[3]
				record.Diacritical = recordCSV[3]
				record.Native = recordCSV[4]
				record.ShortDesc = recordCSV[5]
				record.LongDesc = recordCSV[6]
			}

			// Skip invalid records that have empty term
			if record.Term == "" {
				continue
			}

			if err := index.Index(record.Term, record); err != nil {
				fmt.Println(record)
				return fmt.Errorf("unable to index document %s: %w", record.ID, err)
			}
		}

		log.Printf("Successfully indexed %s branch.", branch)
	}

	log.Println("Successfully indexed all branches")
	return nil
}

func (n *namasteRepository) Find(input string) (*NamasteMatches, error) {
	index, err := bleve.Open("index.bleve")
	if err != nil {
		return nil, fmt.Errorf("unable to open index: %w", err)
	}
	defer index.Close()

	matchQuery := query.NewMatchQuery(input)

	searchRequest := bleve.NewSearchRequest(matchQuery)
	searchRequest.Size = 5 // Get top 5 results
	// We are only concerned with these fields
	searchRequest.Fields = []string{"Type", "Code", "Diacritical", "LongDesc"}

	searchResult, err := index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("unable to search: %w", err)
	}

	matches := make([]namasteMatch, 0, 5)
	for _, hit := range searchResult.Hits {
		typ := hit.Fields["Type"].(string)
		code := hit.Fields["Code"].(string)
		name := hit.Fields["Diacritical"].(string)
		desc := hit.Fields["LongDesc"].(string)

		matches = append(matches, namasteMatch{
			Type: typ,
			ID:   code,
			Name: name,
			Desc: desc,
		})
	}

	return &NamasteMatches{
		matches,
	}, nil
}
