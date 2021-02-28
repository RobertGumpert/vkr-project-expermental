package main

import (
	"encoding/json"
	"errors"
	"fmt"
	githubRequest "github-collector/pckg/github-api/github-request"
	"github-collector/pckg/requests"
	"github-collector/pckg/runtimeinfo"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

type sendResponse func() error
type configuratorResponse func(responses *githubRequest.TaskState, taskCompleted bool) sendResponse

type appService struct {
	config            *config
	client            *http.Client
	repeatedResponses []sendResponse
	GITHUBClient      *githubRequest.GithubClient
}

func NewAppService(config *config) *appService {
	a := &appService{
		config: config,
		client: new(http.Client),
	}
	gitHubClient, err := githubRequest.NewGithubClient(a.config.GithubToken, a.config.CountTasks)
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	a.GITHUBClient = gitHubClient
	go a.repeatResponsesToGithubGate()
	return a
}

func (a *appService) GetReposByURLS(taskKey string, urls []string) error {
	var (
		githubRequests        = make([]githubRequest.Request, 0)
		taskStateChannel      = make(chan *githubRequest.TaskState)
		deferTaskStateChannel = make(chan *githubRequest.TaskState)
	)
	for _, request := range urls {
		githubRequests = append(githubRequests, githubRequest.Request{
			URL:     request,
			TaskKey: taskKey,
			Header: map[string]string{
				"Accept": "application/vnd.github.mercy-preview+json",
			},
		})
	}
	taskGroup, err := a.GITHUBClient.AddGroupRequests(false)
	if err != nil {
		return err
	}
	runTask, noWait, _ := taskGroup(githubRequests, githubRequest.CORE, taskStateChannel, deferTaskStateChannel)
	if noWait {
		go func(responsesChannel, deferResponsesChannel chan *githubRequest.TaskState, runTask githubRequest.RunTask) {
			go runTask()
			a.waitTaskGroupRequests(responsesChannel, deferResponsesChannel, a.configuratorResponseRepositoriesByURL)
			return
		}(taskStateChannel, deferTaskStateChannel, runTask)
	} else {
		go func(responsesChannel, deferResponsesChannel chan *githubRequest.TaskState) {
			a.waitTaskGroupRequests(responsesChannel, deferResponsesChannel, a.configuratorResponseRepositoriesByURL)
			return
		}(taskStateChannel, deferTaskStateChannel)
	}
	return nil
}

func (a *appService) GetRepositoryIssues(taskKey string, url string) error {
	var (
		signalChannel         = make(chan bool)
		getCountIssuesChannel = make(chan *githubRequest.TaskState)
	)
	taskGetCountIssues, err := a.GITHUBClient.AddOneRequest(false)
	if err != nil {
		return err
	}
	reservedTask, err := a.GITHUBClient.AddGroupRequests(true)
	if err != nil {
		return err
	}
	runTaskGetCountIssues, noWait, _ := taskGetCountIssues(githubRequest.Request{
		TaskKey: taskKey,
		URL:     url + "/issues?state=all",
		Header:  nil,
	}, githubRequest.CORE, signalChannel, getCountIssuesChannel)
	go func(signalChannel chan bool) {
		<-signalChannel
		return
	}(signalChannel)
	if noWait {
		go func(taskKey, url string, getCountIssuesChannel chan *githubRequest.TaskState, reservedTaskForIteratePages githubRequest.TaskGroupRequests) {
			go runTaskGetCountIssues()
			a.iteratePagesIssues(
				taskKey,
				url,
				getCountIssuesChannel,
				reservedTaskForIteratePages,
			)
			return
		}(taskKey, url, getCountIssuesChannel, reservedTask)
	} else {
		go func(taskKey, url string, getCountIssuesChannel chan *githubRequest.TaskState, reservedTaskForIteratePages githubRequest.TaskGroupRequests) {
			a.iteratePagesIssues(
				taskKey,
				url,
				getCountIssuesChannel,
				reservedTaskForIteratePages,
			)
			return
		}(taskKey, url, getCountIssuesChannel, reservedTask)
	}
	return nil
}

// Дожидается выполнения задачи на получение
// количества ISSUE в репозитории и
// запускает отложенную задачу TaskGroup
// итерирующуюся по всем страницам содержащим ISSUE.
//
//
func (a *appService) iteratePagesIssues(taskKey, url string, countIssuesChannel chan *githubRequest.TaskState, reservedTask githubRequest.TaskGroupRequests) {
	response := <-countIssuesChannel
	defer response.Responses[0].Response.Body.Close()
	var (
		responsesPagesChannel      = make(chan *githubRequest.TaskState)
		deferResponsesPagesChannel = make(chan *githubRequest.TaskState)
	)
	viewModel := new(ViewModelIssuesList)
	err := json.NewDecoder(response.Responses[0].Response.Body).Decode(viewModel)
	if err != nil {
		runtimeinfo.LogError("non unmarshal list issues [", err, "]")
		var sendResponse = func() error {
			return a.doResponseToGithubGate(
				UpdateTaskStateRepositoryIssues{
					ExecutionTaskStatus: UpdateTaskStateExecutionStatus{
						TaskCompleted: true,
						TaskKey:       taskKey,
					},
					Issues: nil,
				},
				a.config.GithubGateEndpoints.SendResultTaskReposByUlr,
			)
		}
		if err := sendResponse(); err != nil {
			a.repeatedResponses = append(a.repeatedResponses, sendResponse)
		}
	}
	countPages := []ViewModelIssue(*viewModel)[0].Number / 100
	requestsForIssues := make([]githubRequest.Request, 0)
	for page := 0; page < countPages+1; page++ {
		requestsForIssues = append(requestsForIssues, githubRequest.Request{
			TaskKey: taskKey,
			URL:     fmt.Sprintf("%s/issues?state=all&page=%d&per_page=%d", url, page, 100),
			Header:  nil,
		})
	}
	runTask, noWait, _ := reservedTask(
		requestsForIssues,
		githubRequest.CORE,
		responsesPagesChannel,
		deferResponsesPagesChannel,
	)
	if noWait {
		go func(responsesPagesChannel, deferResponsesPagesChannel chan *githubRequest.TaskState, runTask githubRequest.RunTask) {
			go runTask()
			a.waitTaskGroupRequests(responsesPagesChannel, deferResponsesPagesChannel, a.configuratorResponseRepositoryIssues)
			return
		}(responsesPagesChannel, deferResponsesPagesChannel, runTask)
	} else {
		go func(responsesPagesChannel, deferResponsesPagesChannel chan *githubRequest.TaskState) {
			a.waitTaskGroupRequests(responsesPagesChannel, deferResponsesPagesChannel, a.configuratorResponseRepositoryIssues)
			return
		}(responsesPagesChannel, deferResponsesPagesChannel)
	}
}

func (a *appService) sendResponseTaskGroupRequests(responses *githubRequest.TaskState, getResponse configuratorResponse, taskCompleted bool) {
	doResponse := getResponse(responses, taskCompleted)
	if err := doResponse(); err != nil {
		a.repeatedResponses = append(a.repeatedResponses, doResponse)
	}
}

// Для задач типа TaskGroup, дожидается
// первого сообщения из двух каналов.
//
// Если был достигнут RateLimit, то из канала deferResponsesChannel
// получает все уже выполненные запросы, до момента достижения RateLimit,
// и отправляет их сервису github-gate, и запускает горутину,
// которая дожидается завершения всех оставшихся запросов.
//
// Если RateLimit не был достигнут, то из канала responsesChannel
// получает все выполненные запросы и отправляет их сервису github-gate.
//
func (a *appService) waitTaskGroupRequests(taskStateChannel, deferTaskStateChannel chan *githubRequest.TaskState, configuratorResponse configuratorResponse) {
	select {
	case responses := <-deferTaskStateChannel:
		a.sendResponseTaskGroupRequests(responses, configuratorResponse, false)
		go func(deferResponsesChannel chan *githubRequest.TaskState) {
			count := 0
			runtimeinfo.LogInfo("Start")
			for {
				taskState, _ := <-deferResponsesChannel
				count++
				taskCompeted := taskState.ExecutionStatus
				runtimeinfo.LogInfo(count, taskCompeted)
				a.sendResponseTaskGroupRequests(taskState, configuratorResponse, taskCompeted)
				if taskCompeted {
					break
				}
			}
			runtimeinfo.LogInfo("Finish")
			return
		}(deferTaskStateChannel)
	case responses := <-taskStateChannel:
		a.sendResponseTaskGroupRequests(responses, configuratorResponse, true)
	}
}

func (a *appService) configuratorResponseRepositoriesByURL(taskState *githubRequest.TaskState, taskCompleted bool) sendResponse {
	var (
		dispatchedStateTask = UpdateTaskStateReposByURLS{
			ExecutionTaskStatus: UpdateTaskStateExecutionStatus{
				TaskCompleted: taskCompleted,
			},
			Repositories: make([]UpdateTaskStateRepository, 0),
		}
		taskKey string
	)
	for _, response := range taskState.Responses {
		if taskKey == "" {
			taskKey = response.TaskKey
		}
		viewModel := new(ViewModelRepository)
		repository := new(UpdateTaskStateRepository)
		err := json.NewDecoder(response.Response.Body).Decode(viewModel)
		if err != nil {
			repository.URL = response.URL
			repository.Err = err
		}
		if response.Err != nil {
			repository.Err = response.Err
		}
		repository.URL = viewModel.URL
		repository.Topics = viewModel.Topics
		repository.Description = viewModel.Description
		dispatchedStateTask.Repositories = append(dispatchedStateTask.Repositories, *repository)
		if err := response.Response.Body.Close(); err != nil {
			runtimeinfo.LogError("[", response.URL, "]", err)
		}
	}
	dispatchedStateTask.ExecutionTaskStatus.TaskKey = taskKey
	return func() error {
		return a.doResponseToGithubGate(
			&dispatchedStateTask,
			a.config.GithubGateEndpoints.SendResultTaskReposByUlr,
		)
	}
}

func (a *appService) configuratorResponseRepositoryIssues(taskState *githubRequest.TaskState, taskCompleted bool) sendResponse {
	var (
		dispatchedStateTask = UpdateTaskStateRepositoryIssues{
			ExecutionTaskStatus: UpdateTaskStateExecutionStatus{
				TaskCompleted: taskCompleted,
			},
			Issues: make([]UpdateTaskStateIssue, 0),
		}
		taskKey string
	)
	for key, response := range taskState.Responses {
		if response == nil {
			runtimeinfo.LogError("response equals nil :[", key, "];")
			continue
		}
		if response.Response == nil {
			runtimeinfo.LogError("response body equals nil :[", key, "];")
			continue
		}
		if taskKey == "" {
			taskKey = response.TaskKey
		}
		listViewModel := new(ViewModelIssuesList)
		err := json.NewDecoder(response.Response.Body).Decode(listViewModel)
		if err != nil {
			runtimeinfo.LogError("non unmarshal list issues [", err, "]")
			continue
		}
		for _, issue := range []ViewModelIssue(*listViewModel) {
			dispatchedStateTask.Issues = append(dispatchedStateTask.Issues, UpdateTaskStateIssue{
				Number: issue.Number,
				URL:    issue.URL,
				Title:  issue.Title,
				State:  issue.State,
				Body:   issue.Body,
				Err:    response.Err,
			})
		}
		if err := response.Response.Body.Close(); err != nil {
			runtimeinfo.LogError("[", response.URL, "]", err)
		}
	}
	dispatchedStateTask.ExecutionTaskStatus.TaskKey = taskKey
	return func() error {
		return a.doResponseToGithubGate(
			dispatchedStateTask,
			a.config.GithubGateEndpoints.SendResultTaskIssueRepo,
		)
	}
}

func (a *appService) repeatResponsesToGithubGate() {
	for {
		runtime.Gosched()
		runtimeinfo.LogInfo("REPEATED TASK RESPONSES...")
		for index, repeatedResponse := range a.repeatedResponses {
			if err := repeatedResponse(); err == nil {
				a.repeatedResponses = append(a.repeatedResponses[:index], a.repeatedResponses[index+1:]...)
			}
		}
		runtimeinfo.LogInfo("REPEATED TASK RESPONSES SLEEP...")
		time.Sleep(10 * time.Minute)
	}
}

func (a *appService) doResponseToGithubGate(body interface{}, endpoint string) error {
	url := fmt.Sprintf("%s%s", a.config.GithubGateAddress, endpoint)
	response, err := requests.POST(
		a.client,
		url,
		nil,
		body,
	)
	if err != nil {
		runtimeinfo.LogError("[", url, "]", err)
		return err
	}
	if response.StatusCode != http.StatusOK {
		err := errors.New("[" + url + "] status: " + strconv.Itoa(response.StatusCode))
		runtimeinfo.LogError(err)
		return err
	}
	runtimeinfo.LogInfo("[", url, "] OK.")
	return nil
}
