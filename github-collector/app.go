package main

import (
	githubRequest "github-collector/pckg/github-api/github-request"
)

type AppService struct {
	GITHUBClient *githubRequest.GithubClient
}

func NewAppService(GITHUBClient *githubRequest.GithubClient) *AppService {
	return &AppService{GITHUBClient: GITHUBClient}
}

func (a *AppService) GetRepos(urls []string) {
	var (
		requests  = make([]githubRequest.Request, 0)
		responses map[string]*githubRequest.Response
		mainChannelResponses,
		lazyChannelResponses = make(chan map[string]*githubRequest.Response),
			make(chan map[string]*githubRequest.Response)
	)
	for _, request := range urls {
		requests = append(requests, githubRequest.Request{
			URL: request,
			Header: map[string]string{
				"Accept": "application/vnd.github.mercy-preview+json",
			},
		})
	}
	go GITHUBClient.Requests(
		requests,
		githubRequest.CORE,
		mainChannelResponses,
		lazyChannelResponses,
	)
	select {
	case responses = <-mainChannelResponses:

	case responses = <-lazyChannelResponses:

	}
}

func (a *AppService) unmarshalRepositories(responses map[string]*githubRequest.Response) {
	for key, response := range responses {

	}
}
