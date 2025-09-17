package dto

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type MatchingPV struct {
	PropertyID string `json:"propertyId"`
	Label      string `json:"label"`
}

type DestinationEntity struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	TheCode     string       `json:"theCode"`
	MatchingPVs []MatchingPV `json:"matchingPVs"`
}

type SearchResponse struct {
	DestinationEntities []DestinationEntity `json:"destinationEntities"`
}
