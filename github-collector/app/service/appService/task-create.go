package appService

import (
	"github-collector/app/service/githubApiService"
	"strings"
)

func (service *AppService) createTaskRepositoriesDescriptions(jsonModel *JsonCreateTaskRepositoriesDescriptions) (err error) {
	var (
		requests                        = make([]githubApiService.Request, 0)
		channelResponsesBeforeRateLimit = make(chan *githubApiService.TaskState)
		channelResponsesAfterRateLimit  = make(chan *githubApiService.TaskState)
	)
	for _, repository := range jsonModel.Repositories {
		url := strings.Join(
			[]string{
				"https://api.github.com/repos",
				repository.Owner,
				repository.Name,
			},
			"/",
		)
		requests = append(requests, githubApiService.Request{
			TaskKey: jsonModel.TaskKey,
			URL:     url,
			Header: map[string]string{
				"Accept": "application/vnd.github.mercy-preview+json",
			},
		})
	}
	functionInsertGroupRequests, queueIsFilled := service.GithubClient.MakeFunctionInsertGroupRequests(false)
	if queueIsFilled {
		return ErrorQueueIsFilled
	}
	taskLaunchFunction, isRunTaskNow, _ := functionInsertGroupRequests(
		requests,
		githubApiService.CORE,
		channelResponsesBeforeRateLimit,
		channelResponsesAfterRateLimit,
	)
	go service.waitTaskRepositoriesDescriptions(
		isRunTaskNow,
		taskLaunchFunction,
		channelResponsesBeforeRateLimit,
		channelResponsesAfterRateLimit,
	)
	return nil
}

func (service *AppService) createTaskRepositoryIssues(jsonModel *JsonCreateTaskRepositoryIssues) (err error) {
	var (
		requestForFindCountPages     githubApiService.Request
		channelNotificationRateLimit = make(chan bool)
		channelGettingTaskState      = make(chan *githubApiService.TaskState)
		repositoryIssuesURL          = strings.Join(
			[]string{
				"https://api.github.com/repos",
				jsonModel.Repository.Owner,
				jsonModel.Repository.Name,
				"issues?state=all",
			},
			"/",
		)
	)
	requestForFindCountPages = githubApiService.Request{
		TaskKey: jsonModel.TaskKey,
		URL:     repositoryIssuesURL,
		Header:  nil,
	}
	functionInsertOneRequest, queueIsFilled := service.GithubClient.MakeFunctionInsertOneRequest(false)
	if queueIsFilled {
		return ErrorQueueIsFilled
	}
	functionInsertGroupRequests, queueIsFilled := service.GithubClient.MakeFunctionInsertGroupRequests(true)
	if queueIsFilled {
		return ErrorQueueIsFilled
	}
	taskLaunchFunction, isRunTaskNow, _ := functionInsertOneRequest(
		requestForFindCountPages,
		githubApiService.CORE,
		channelNotificationRateLimit,
		channelGettingTaskState,
	)
	go service.waitTaskRepositoryIssues(
		isRunTaskNow,
		taskLaunchFunction,
		channelNotificationRateLimit,
		channelGettingTaskState,
		functionInsertGroupRequests,
		repositoryIssuesURL,
		jsonModel.TaskKey,
	)
	return nil
}

func (service *AppService) createTaskRepositoriesByKeyWord(jsonModel *JsonCreateTaskRepositoriesByKeyWord) (err error) {
	var (
		request                      githubApiService.Request
		channelNotificationRateLimit = make(chan bool)
		channelGettingTaskState      = make(chan *githubApiService.TaskState)
		repositoryIssuesURL          = strings.Join(
			[]string{
				"https://api.github.com/search/repositories?q=topic:",
				jsonModel.KeyWord,
				"&per_page=30",
			},
			"",
		)
	)
	request = githubApiService.Request{
		TaskKey: jsonModel.TaskKey,
		URL:     repositoryIssuesURL,
		Header: map[string]string{
			"Accept": "application/vnd.github.mercy-preview+json",
		},
	}
	functionInsertOneRequest, queueIsFilled := service.GithubClient.MakeFunctionInsertOneRequest(false)
	if queueIsFilled {
		return ErrorQueueIsFilled
	}
	taskLaunchFunction, isRunTaskNow, _ := functionInsertOneRequest(
		request,
		githubApiService.SEARCH,
		channelNotificationRateLimit,
		channelGettingTaskState,
	)
	go service.waitTaskRepositoriesByKeyWord(
		isRunTaskNow,
		taskLaunchFunction,
		channelNotificationRateLimit,
		channelGettingTaskState,
	)
	return nil
}