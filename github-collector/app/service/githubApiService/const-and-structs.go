package githubApiService

import "net/http"

type GitHubLevelAPI uint64

const (
	CORE                GitHubLevelAPI = 0
	SEARCH              GitHubLevelAPI = 1
	maxCoreRequests     uint64         = 5000
	maxSearchRequests   uint64         = 30
	limitNumberAttempts int            = 5
	authURL                            = "https://api.github.com/user"
	rateLimitURL                       = "https://api.github.com/rate_limit"
)

type TaskState struct {
	TaskKey       string
	TaskCompleted bool
	Responses     []*Response
}

type Response struct {
	TaskKey  string
	URL      string
	Response *http.Response
	Err      error
}

type Request struct {
	TaskKey             string
	URL                 string
	Header              map[string]string
	numberSpentAttempts int
}
