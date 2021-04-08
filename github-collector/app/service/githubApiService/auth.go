package githubApiService

import (
	"errors"
	"github-collector/pckg/requests"
	"net/http"
)

func (c *GithubClient) auth() error {
	response, err := requests.GET(c.client, authURL, map[string]string{
		"Authorization": c.token,
	})
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("Status code not 200. ")
	}
	return nil
}

func (c *GithubClient) addAuthHeader(header map[string]string) map[string]string {
	if header == nil && c.isAuth {
		header = map[string]string{
			"Authorization": c.token,
		}
	}
	if header != nil && c.isAuth {
		header["Authorization"] = c.token
	}
	return header
}
