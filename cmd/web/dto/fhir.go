package dto

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
