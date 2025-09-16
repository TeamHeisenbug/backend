package dto

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type DestinationEntity struct {
	Title   string `json:"title"`
	TheCode string `json:"theCode"`
}

type SearchResponse struct {
	DestinationEntities []DestinationEntity `json:"destinationEntities"`
}
