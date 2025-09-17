package dto

import "time"

type ValueSet struct {
	ResourceType string    `json:"resourceType"` // ValueSet
	ID           string    `json:"id"`           // autocomplete-results
	Status       string    `json:"status"`       // active
	Expansion    Expansion `json:"expansion"`
}

type Expansion struct {
	Identifier string    `json:"identifier"` // link to autocomplete endpoint
	Timestamp  time.Time `json:"timestamp"`
	Total      int       `json:"total"`  // 2
	Offset     int       `json:"offset"` // 0
	Contains   []Contain `json:"contains"`
}

type Contain struct {
	System    string    `json:"system"`  // link to namaste/icd code system page
	Code      string    `json:"code"`    // code
	Display   string    `json:"display"` // term
	Extension Extension `json:"extension"`
}

type Extension struct {
	URL         string `json:"url"`         // link to structure defintion
	ValueString string `json:"valueString"` // NAMASTE/ICD
}

type CodeSystem struct {
	ResourceType string    `json:"resourceType"` // CodeSystem
	ID           string    `json:"id"`           // NAMASTE
	URL          string    `json:"url"`
	Version      string    `json:"version"`
	Name         string    `json:"name"`    // NAMASTE Codes
	Status       string    `json:"status"`  // active
	Content      string    `json:"content"` // Complete
	Concept      []Concept `json:"concept"`
}

type Concept struct {
	Code       string     `json:"code"`       // code
	Display    string     `json:"display"`    // Term
	Definition string     `json:"definition"` // longDesc
	Property   []Property `json:"property"`
}

type Property struct {
	Code        string `json:"code"`        // type
	ValueString string `json:"valueString"` // ayurveda/siddha/unani
}
