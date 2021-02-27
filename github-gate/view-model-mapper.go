package main


type ViewModelRepository struct {
	URL         string   `json:"url"`
	Topics      []string `json:"topics"`
	Description string   `json:"about"`
}