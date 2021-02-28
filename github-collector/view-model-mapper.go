package main

type ViewModelIssue struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	State  string   `json:"state"`
	Body   string `json:"body"`
}

type ViewModelIssuesList []ViewModelIssue

type ViewModelRepository struct {
	URL         string   `json:"url"`
	Topics      []string `json:"topics"`
	Description string   `json:"description"`
}
