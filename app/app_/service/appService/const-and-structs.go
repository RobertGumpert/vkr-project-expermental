package appService

type JsonTaskFindNearestRepositories struct {
	Keyword string `json:"keyword"`
	Name    string `json:"name"`
	Owner   string `json:"owner"`
	Email   string `json:"email"`
}
