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
type configuratorResponse func(responses map[string]*githubRequest.Response, taskCompleted bool) sendResponse

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
		githubRequests = make([]githubRequest.Request, 0)
		responsesChannel,
		deferResponsesChannel = make(chan map[string]*githubRequest.Response),
			make(chan map[string]*githubRequest.Response)
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
	runTask, noWait, _ := taskGroup(githubRequests, githubRequest.CORE, responsesChannel, deferResponsesChannel)
	if noWait {
		go func(responsesChannel, deferResponsesChannel chan map[string]*githubRequest.Response, runTask githubRequest.RunTask) {
			go runTask()
			a.waitTaskGroupRequests(responsesChannel, deferResponsesChannel, a.configuratorResponseGroupRequests)
			return
		}(responsesChannel, deferResponsesChannel, runTask)
	} else {
		go func(responsesChannel, deferResponsesChannel chan map[string]*githubRequest.Response) {
			a.waitTaskGroupRequests(responsesChannel, deferResponsesChannel, a.configuratorResponseGroupRequests)
			return
		}(responsesChannel, deferResponsesChannel)
	}
	return nil
}

func (a *appService) GetRepositoryIssues(taskKey string, url string) error {
	var (
		signalChannel         = make(chan bool)
		getCountIssuesChannel = make(chan *githubRequest.Response)
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
	if noWait {
		go func(taskKey, url string, getCountIssuesChannel chan *githubRequest.Response, reservedTaskForIteratePages githubRequest.TaskGroupRequests) {
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
		go func(taskKey, url string, getCountIssuesChannel chan *githubRequest.Response, reservedTaskForIteratePages githubRequest.TaskGroupRequests) {
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
func (a *appService) iteratePagesIssues(taskKey, url string, countIssuesChannel chan *githubRequest.Response, reservedTask githubRequest.TaskGroupRequests) {
	response := <-countIssuesChannel
	defer response.Response.Body.Close()
	var (
		responsesPagesChannel      = make(chan map[string]*githubRequest.Response)
		deferResponsesPagesChannel = make(chan map[string]*githubRequest.Response)
	)
	viewModel := new(ViewModelIssuesList)
	err := json.NewDecoder(response.Response.Body).Decode(viewModel)
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
		go func(responsesPagesChannel, deferResponsesPagesChannel chan map[string]*githubRequest.Response, runTask githubRequest.RunTask) {
			go runTask()
			a.waitTaskGroupRequests(responsesPagesChannel, deferResponsesPagesChannel, a.configuratorResponseRepositoryIssues)
			return
		}(responsesPagesChannel, deferResponsesPagesChannel, runTask)
	} else {
		go func(responsesPagesChannel, deferResponsesPagesChannel chan map[string]*githubRequest.Response) {
			a.waitTaskGroupRequests(responsesPagesChannel, deferResponsesPagesChannel, a.configuratorResponseRepositoryIssues)
			return
		}(responsesPagesChannel, deferResponsesPagesChannel)
	}
}

func (a *appService) sendResponseTaskGroupRequests(responses map[string]*githubRequest.Response, getResponse configuratorResponse, taskCompleted bool) {
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
func (a *appService) waitTaskGroupRequests(responsesChannel, deferResponsesChannel chan map[string]*githubRequest.Response, configuratorResponse configuratorResponse) {
	select {
	case responses := <-deferResponsesChannel:
		a.sendResponseTaskGroupRequests(responses, configuratorResponse, false)
		go func(deferResponsesChannel chan map[string]*githubRequest.Response) {
			responses := <-deferResponsesChannel
			a.sendResponseTaskGroupRequests(responses, configuratorResponse, true)
			return
		}(deferResponsesChannel)
	case responses := <-responsesChannel:
		a.sendResponseTaskGroupRequests(responses, configuratorResponse, true)
	}
}

func (a *appService) configuratorResponseGroupRequests(responses map[string]*githubRequest.Response, taskCompleted bool) sendResponse {
	var (
		updateTask = UpdateTaskStateReposByURLS{
			ExecutionTaskStatus: UpdateTaskStateExecutionStatus{
				TaskCompleted: taskCompleted,
			},
			Repositories: make([]UpdateTaskStateRepository, 0),
		}
		taskKey string
	)
	for _, response := range responses {
		if taskKey == "" {
			taskKey = response.TaskKey
		}
		viewModelRepository := new(ViewModelRepository)
		updateTaskRepository := new(UpdateTaskStateRepository)
		err := json.NewDecoder(response.Response.Body).Decode(viewModelRepository)
		if err != nil {
			updateTaskRepository.URL = response.URL
			updateTaskRepository.Err = err
		}
		if response.Err != nil {
			updateTaskRepository.Err = response.Err
		}
		updateTaskRepository.URL = viewModelRepository.URL
		updateTaskRepository.Topics = viewModelRepository.Topics
		updateTaskRepository.Description = viewModelRepository.Description
		updateTask.Repositories = append(updateTask.Repositories, *updateTaskRepository)
		if err := response.Response.Body.Close(); err != nil {
			runtimeinfo.LogError("[", response.URL, "]", err)
		}
	}
	updateTask.ExecutionTaskStatus.TaskKey = taskKey
	return func() error {
		return a.doResponseToGithubGate(
			&updateTask,
			a.config.GithubGateEndpoints.SendResultTaskReposByUlr,
		)
	}
}

func (a *appService) configuratorResponseRepositoryIssues(responses map[string]*githubRequest.Response, taskCompleted bool) sendResponse {
	var (
		updateTask = UpdateTaskStateRepositoryIssues{
			ExecutionTaskStatus: UpdateTaskStateExecutionStatus{
				TaskCompleted: taskCompleted,
			},
			Issues: make([]UpdateTaskStateIssue, 0),
		}
		taskKey string
	)
	for _, response := range responses {
		if taskKey == "" {
			taskKey = response.TaskKey
		}
		viewModelIssueList := new(ViewModelIssuesList)
		err := json.NewDecoder(response.Response.Body).Decode(viewModelIssueList)
		if err != nil {
			runtimeinfo.LogError("non unmarshal list issues [", err, "]")
			continue
		}
		for _, issue := range []ViewModelIssue(*viewModelIssueList) {
			updateTask.Issues = append(updateTask.Issues, UpdateTaskStateIssue{
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
	updateTask.ExecutionTaskStatus.TaskKey = taskKey
	return func() error {
		return a.doResponseToGithubGate(
			updateTask,
			a.config.GithubGateEndpoints.SendResultTaskReposByUlr,
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
