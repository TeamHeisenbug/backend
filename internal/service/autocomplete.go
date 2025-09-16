package service

import (
	"backend/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"google.golang.org/genai"
)

type ICD struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type Namaste struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type Disease struct {
	ICD     ICD     `json:"icd"`
	Namaste Namaste `json:"namaste"`
}

type Matches struct {
	Diseases []Disease `json:"diseases"`
}

type AutoComplete interface {
	Update() error
	Find(ctx context.Context, input string) (*Matches, error)
}

type autoComplete struct {
	genaiClient       *genai.Client
	icdRepository     repository.ICDRepository
	namasteRepository repository.NamasteRepository
}

const prompt = `
Here is the ICDRepository response: %s
Here is the NamasteRepository response: %s
I want you to carefully match the corresponding diseases from both responses according to the similarity of their descriptions.
Return the final output strictly in the following JSON format only:
{
  "diseases": [
    {
      "icd": {
        "id": "string",
        "name": "string"
      },
      "namaste": {
        "type": "string",
        "id": "string",
        "name": "string",
        "desc": "string"
      }
    }
  ]
}
`

// Find implements AutoComplete.
func (a *autoComplete) Find(ctx context.Context, input string) (*Matches, error) {
	icdMatches, err := a.icdRepository.Find(input)
	if err != nil {
		return nil, err
	}

	namasteMatches, err := a.namasteRepository.Find(input)
	if err != nil {
		return nil, err
	}

	log.Println("icdMatches:")
	log.Println(icdMatches)
	log.Print("namasteMatches:")
	log.Println(namasteMatches)

	genaiResponse, err := a.genaiClient.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(fmt.Sprintf(prompt, icdMatches, namasteMatches)), nil)
	if err != nil {
		return nil, err
	}

	resultLines := strings.Split(genaiResponse.Text(), "\n")

	fmt.Println(genaiResponse.Text())

	result := ""

	for i, line := range resultLines {
		if i == 0 || i == len(resultLines)-1 {
			continue
		}

		result += line + "\n"
	}

	fmt.Println(result)

	var matches Matches
	if err := json.Unmarshal([]byte(result), &matches); err != nil {
		return nil, fmt.Errorf("unable to decode genai response: %w", err)
	}

	return &matches, nil
}

func (a *autoComplete) Update() error {
	a.namasteRepository.CreateIndex("index.bleve")
	return nil
}

func NewAutoComplete(genaiClient *genai.Client, icdRepository repository.ICDRepository, namasteRepository repository.NamasteRepository) AutoComplete {
	return &autoComplete{
		genaiClient:       genaiClient,
		icdRepository:     icdRepository,
		namasteRepository: namasteRepository,
	}
}
