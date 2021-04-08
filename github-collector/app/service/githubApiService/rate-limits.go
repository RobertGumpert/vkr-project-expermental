package githubApiService


import (
	"errors"
	"github-collector/pckg/requests"
	"github-collector/pckg/runtimeinfo"
	"time"
)

type resourcesRateLimitsJSON struct {
	Resources struct {
		Core struct {
			Used  int64 `json:"used"`
			Reset int64 `json:"reset"`
		} `json:"core"`
		Search struct {
			Used  int64 `json:"used"`
			Reset int64 `json:"reset"`
		} `json:"search"`
	} `json:"resources"`
	Rate struct {
		Used  int64 `json:"used"`
		Reset int64 `json:"reset"`
	} `json:"rate"`
}

func (c *GithubClient) getRateLimit() (*resourcesRateLimitsJSON, error) {
	response, err := requests.GET(c.client, rateLimitURL, c.addAuthHeader(nil))
	if err != nil {
		return nil, err
	}
	var rate *resourcesRateLimitsJSON
	if err := requests.Deserialize(&rate, response); err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New("Status code not 200. ")
	}
	if err := response.Body.Close(); err != nil {
		runtimeinfo.LogError(err)
	}
	return rate, nil
}

func (c *GithubClient) freezeClient(reset int64) {
	timeNow := time.Now()
	timeReset := time.Unix(reset, int64(0))
	var when time.Duration
	if timeNow.After(timeReset) {
		when = timeNow.Sub(timeReset)
	} else {
		when = timeReset.Sub(timeNow)
	}
	runtimeinfo.LogInfo("CLIENT FREEZE ON ", when, "...")
	time.Sleep(when)
	runtimeinfo.LogInfo("CLIENT UNFREEZE.")
}

