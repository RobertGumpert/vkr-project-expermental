package main

import (
	"encoding/json"
	githubRequest "github-collector/pckg/github-api/github-request"
	"net/http"
)

type AppService struct {
	client        *http.Client
	GitHubGateURL string
	GITHUBClient  *githubRequest.GithubClient
}

type sendToGithubGate func(responses map[string]*githubRequest.Response, taskCompleted bool)

func NewAppService(GITHUBClient *githubRequest.GithubClient, gitHubGateURL string) *AppService {
	return &AppService{
		client:        new(http.Client),
		GitHubGateURL: gitHubGateURL,
		GITHUBClient:  GITHUBClient,
	}
}

func (a *AppService) GetReposByURLS(taskKey string, urls []string) error {
	var (
		requests = make([]githubRequest.Request, 0)
		responsesChannel,
		deferResponsesChannel = make(chan map[string]*githubRequest.Response),
			make(chan map[string]*githubRequest.Response)
	)
	for _, request := range urls {
		requests = append(requests, githubRequest.Request{
			URL:     request,
			TaskKey: taskKey,
			Header: map[string]string{
				"Accept": "application/vnd.github.mercy-preview+json",
			},
		})
	}
	taskGroup, err := a.GITHUBClient.AddGroupRequests()
	if err != nil {
		return err
	}
	event, noWait := taskGroup(requests, githubRequest.CORE, responsesChannel, deferResponsesChannel)
	if noWait {
		go func(responsesChannel, deferResponsesChannel chan map[string]*githubRequest.Response, event githubRequest.Event) {
			go event()
			a.waitGroupRequests(responsesChannel, deferResponsesChannel, a.sendRepositoriesToGithubGate)
			return
		}(responsesChannel, deferResponsesChannel, event)
	} else {
		go func(responsesChannel, deferResponsesChannel chan map[string]*githubRequest.Response) {
			a.waitGroupRequests(responsesChannel, deferResponsesChannel, a.sendRepositoriesToGithubGate)
			return
		}(responsesChannel, deferResponsesChannel)
	}
	return nil
}

func (a *AppService) waitGroupRequests(responsesChannel, deferResponsesChannel chan map[string]*githubRequest.Response, send sendToGithubGate) {
	select {
	case responses := <-deferResponsesChannel:
		send(responses, false)
		go func(deferResponsesChannel chan map[string]*githubRequest.Response) {
			responses := <-deferResponsesChannel
			send(responses, true)
			return
		}(deferResponsesChannel)
	case responses := <-responsesChannel:
		send(responses, true)
	}
}

func (a *AppService) sendRepositoriesToGithubGate(responses map[string]*githubRequest.Response, taskCompleted bool) {
	var (
		repositories        = make([]JSONRepository, 0)
		taskExecutionStatus = JSONExecutionTaskStatus{}
	)
	taskExecutionStatus.TaskCompleted = taskCompleted
	for _, value := range responses {
		taskExecutionStatus.TaskKey = value.TaskKey
		response := new(JSONRepository)
		err := json.NewDecoder(value.Response.Body).Decode(response)
		if err != nil {
			response.URL = value.URL
		}
		response.URL = value.URL
		repositories = append(repositories, *response)
	}
}
