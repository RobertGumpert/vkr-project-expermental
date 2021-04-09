package appService

import (
	"encoding/json"
	"fmt"
	"github-collector/app/service/githubApiService"
	"github-collector/pckg/runtimeinfo"
)

func (service *AppService) waitTaskRepositoriesDescriptions(
	isRunTaskNow githubApiService.IsRunTaskNow,
	taskLaunchFunction githubApiService.TaskLaunchFunction,
	channelResponsesBeforeRateLimit,
	channelResponsesAfterRateLimit chan *githubApiService.TaskState,
) {
	if isRunTaskNow {
		go taskLaunchFunction()
	}
	select {
	case allCompletedRequests := <-channelResponsesBeforeRateLimit:
		service.sendTaskRepositoriesDescriptions(allCompletedRequests)
		break
	case partOfCompletedRequests := <-channelResponsesAfterRateLimit:
		service.sendTaskRepositoriesDescriptions(partOfCompletedRequests)
		go func(channelResponsesAfterRateLimit chan *githubApiService.TaskState) {
			countPartOfRequests := 0
			runtimeinfo.LogInfo("Start")
			for {
				nextPartOfCompletedRequests, _ := <-channelResponsesAfterRateLimit
				countPartOfRequests++
				runtimeinfo.LogInfo("PART OF REQUESTS: [", countPartOfRequests, "] IS COMPLETED: [", nextPartOfCompletedRequests.TaskCompleted, "]")
				service.sendTaskRepositoriesDescriptions(nextPartOfCompletedRequests)
				if nextPartOfCompletedRequests.TaskCompleted {
					break
				}
			}
			runtimeinfo.LogInfo("Finish")
			return
		}(channelResponsesAfterRateLimit)
		break
	}
	return
}

func (service *AppService) waitTaskRepositoryIssues(
	isRunTaskNow githubApiService.IsRunTaskNow,
	taskLaunchFunction githubApiService.TaskLaunchFunction,
	channelNotificationRateLimit chan bool,
	channelGettingTaskState chan *githubApiService.TaskState,
	functionInsertGroupRequests githubApiService.FunctionInsertGroupRequests,
	repositoryIssuesURL string,
	taskKey string,
) {
	var (
		countPages                      int
		requests                        = make([]githubApiService.Request, 0)
		countPagesDataModel             = new(ListIssuesDataModel)
		channelResponsesBeforeRateLimit = make(chan *githubApiService.TaskState)
		channelResponsesAfterRateLimit  = make(chan *githubApiService.TaskState)
	)
	if isRunTaskNow {
		go taskLaunchFunction()
	}
	go func(channelNotificationRateLimit chan bool) {
		<-channelNotificationRateLimit
		return
	}(channelNotificationRateLimit)
	countPagesRequest := <-channelGettingTaskState
	if countPagesRequest.Responses[0].Response == nil {
		countPagesRequest.TaskCompleted = true
		service.sendTaskRepositoryIssues(countPagesRequest)
		return
	}
	err := json.NewDecoder(countPagesRequest.Responses[0].Response.Body).Decode(countPagesDataModel)
	if err != nil {
		countPagesRequest.TaskCompleted = true
		service.sendTaskRepositoryIssues(countPagesRequest)
		return
	}
	countPages = []IssueDataModel(*countPagesDataModel)[0].Number / 100
	for pageNumber := 0; pageNumber < countPages+1; pageNumber++ {
		pageUrl := fmt.Sprintf(
			"%s&page=%d&per_page=%d",
			repositoryIssuesURL,
			pageNumber,
			100,
		)
		request := githubApiService.Request{
			TaskKey: taskKey,
			URL:     pageUrl,
			Header:  nil,
		}
		requests = append(requests, request)
	}
	groupTaskLaunchFunction, isRunGroupTaskNow, _ := functionInsertGroupRequests(
		requests,
		githubApiService.CORE,
		channelResponsesBeforeRateLimit,
		channelResponsesAfterRateLimit,
	)
	go service.waitPaginationRepositoryIssues(
		isRunGroupTaskNow,
		groupTaskLaunchFunction,
		channelResponsesBeforeRateLimit,
		channelResponsesAfterRateLimit,
	)
	return
}

func (service *AppService) waitPaginationRepositoryIssues(
	isRunTaskNow githubApiService.IsRunTaskNow,
	taskLaunchFunction githubApiService.TaskLaunchFunction,
	channelResponsesBeforeRateLimit,
	channelResponsesAfterRateLimit chan *githubApiService.TaskState,
) {
	if isRunTaskNow {
		go taskLaunchFunction()
	}
	select {
	case allCompletedRequests := <-channelResponsesBeforeRateLimit:
		service.sendTaskRepositoryIssues(allCompletedRequests)
		break
	case partOfCompletedRequests := <-channelResponsesAfterRateLimit:
		service.sendTaskRepositoryIssues(partOfCompletedRequests)
		go func(channelResponsesAfterRateLimit chan *githubApiService.TaskState) {
			countPartOfRequests := 0
			runtimeinfo.LogInfo("Start")
			for {
				nextPartOfCompletedRequests, _ := <-channelResponsesAfterRateLimit
				countPartOfRequests++
				runtimeinfo.LogInfo("PART OF REQUESTS: [", countPartOfRequests, "]; IS COMPLETED: [", nextPartOfCompletedRequests.TaskCompleted, "]; TASK: [", nextPartOfCompletedRequests.TaskKey, "]")
				service.sendTaskRepositoryIssues(nextPartOfCompletedRequests)
				if nextPartOfCompletedRequests.TaskCompleted {
					runtimeinfo.LogInfo("Finish")
					break
				}
			}
			return
		}(channelResponsesAfterRateLimit)
		break
	}
	return
}

func (service *AppService) waitTaskRepositoriesByKeyWord(
	isRunTaskNow githubApiService.IsRunTaskNow,
	taskLaunchFunction githubApiService.TaskLaunchFunction,
	channelNotificationRateLimit chan bool,
	channelGettingTaskState chan *githubApiService.TaskState,
) {
	if isRunTaskNow {
		go taskLaunchFunction()
	}
	select {
	case allCompletedRequests := <-channelGettingTaskState:
		service.sendTaskRepositoriesByKeyWord(allCompletedRequests)
		break
	case _ = <-channelNotificationRateLimit:
		go func(channelGettingTaskState chan *githubApiService.TaskState) {
			allCompletedRequests := <-channelGettingTaskState
			service.sendTaskRepositoriesByKeyWord(allCompletedRequests)
			return
		}(channelGettingTaskState)
		break
	}
	return
}
