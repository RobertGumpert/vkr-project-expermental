package main

import (
	"errors"
	"fmt"
	"github-gate/pckg/requests"
	"github-gate/pckg/runtimeinfo"
	cmap "github.com/streamrail/concurrent-map"
	"net/http"
	"strconv"
	"strings"
)

type taskType uint

const (
	TypeTaskGetRepositoriesByURL taskType = 0
	TypeTaskGetRepositoryIssue   taskType = 1
)

type task struct {
	taskType         taskType
	TaskKey          string
	ExecutionStatus  bool
	About            string
	CollectorAddress string
	Results          interface{}
	//
	DeferSendTask     bool
	TaskSend          bool
	SendBody          interface{}
	CollectorEndpoint string
	//
}

type appService struct {
	config       *config
	client       *http.Client
	tasks        *cmap.ConcurrentMap
	tasksChannel chan string
}

func NewAppService(config *config) *appService {
	tasks := cmap.New()
	a := &appService{
		config:       config,
		client:       new(http.Client),
		tasks:        &tasks,
		tasksChannel: make(chan string, config.CountTask),
	}
	go a.scanTasks()
	return a
}

func (a *appService) GetFreeCollector() int {
	for index, url := range a.config.GithubCollectorsAddresses {
		url = url + "/get/state"
		response, err := requests.GET(
			a.client,
			url,
			nil,
		)
		if err != nil {
			runtimeinfo.LogError("[", url, "] error: ", err)
			continue
		}
		if response.StatusCode == http.StatusOK {
			runtimeinfo.LogInfo("[", url, "] status: ", response.StatusCode)
			return index
		} else {
			runtimeinfo.LogError("[", url, "] status: ", response.StatusCode)
		}
	}
	return -1
}

func (a *appService) CreateTaskRepositoriesByURL(createTaskModel *CreateTaskRepoByURLS) error {
	collector := a.GetFreeCollector()
	if collector == -1 {
		err := errors.New("Non free GitHub collectors. ")
		runtimeinfo.LogError(err)
		return err
	}
	repositories := make([]*ViewModelRepository, 0)
	task := &task{
		taskType:         TypeTaskGetRepositoriesByURL,
		ExecutionStatus:  false,
		About:            strings.Join(createTaskModel.Repositories, ", "),
		TaskKey:          fmt.Sprintf("key [%d]", a.tasks.Count()+1),
		CollectorAddress: a.config.GithubCollectorsAddresses[collector],
		Results:          &repositories,
	}
	a.tasks.Set(task.TaskKey, task)
	url := fmt.Sprintf("%s%s", task.CollectorAddress, "/get/repos/by/url")
	response, err := requests.POST(a.client, url, nil, &SendTaskReposByURL{
		TaskKey: task.TaskKey,
		URLS:    createTaskModel.Repositories,
	})
	if err != nil {
		runtimeinfo.LogError("request on create task [", url, "] error: ", err)
		return err
	}
	status := "request on create task [" + url + "] status: " + strconv.Itoa(response.StatusCode)
	if response.StatusCode == http.StatusOK {
		runtimeinfo.LogInfo(status)
	} else {
		err := errors.New(status)
		runtimeinfo.LogError(err)
		return err
	}
	return nil
}

func (a *appService) CreateTaskGetRepositoriesIssues(urls []string) error {
	for _, url := range urls {
		taskKey := fmt.Sprintf("defer key [%d]", a.tasks.Count()+1)
		collectorEndpoint := "/get/repos/issues"
		sendBody := &SendTaskRepositoryIssues{
			TaskKey: taskKey,
			URL:     url,
		}
		task := &task{
			taskType:          TypeTaskGetRepositoryIssue,
			ExecutionStatus:   false,
			TaskSend:          false,
			DeferSendTask:     true,
			About:             "issues for repository [" + url + "]",
			TaskKey:           taskKey,
			SendBody:          sendBody,
			CollectorEndpoint: collectorEndpoint,
			Results:           0,
		}
		indexCollector := a.GetFreeCollector()
		if indexCollector != -1 {
			urlCollector := fmt.Sprintf("%s%s", a.config.GithubCollectorsAddresses[indexCollector], collectorEndpoint)
			response, err := requests.POST(a.client, urlCollector, nil, task.SendBody)
			if err == nil && response.StatusCode == http.StatusOK {
				task.TaskSend = true
				task.CollectorAddress = a.config.GithubCollectorsAddresses[indexCollector]
				runtimeinfo.LogInfo("run defer task by key :[", task.TaskKey, "]; on free collector.")
			}
		} else {
			runtimeinfo.LogInfo("defer task by key :[", task.TaskKey, "];")
		}
		a.tasks.Set(task.TaskKey, task)
	}
	return nil
}

func (a *appService) UpdateStateTaskRepositoriesByURL(updateTaskState *UpdateTaskReposByURLS) error {
	key := updateTaskState.ExecutionTaskStatus.TaskKey
	if value, exist := a.tasks.Get(key); !exist {
		err := errors.New("task with key [" + key + "] isn't exist ")
		runtimeinfo.LogError(err)
		return err
	} else {
		task := value.(*task)
		task.ExecutionStatus = updateTaskState.ExecutionTaskStatus.TaskCompleted
		if task.ExecutionStatus {
			a.tasksChannel <- task.TaskKey
		}
		runtimeinfo.LogInfo("task with key :[", task.TaskKey, "] is competed ;[", task.ExecutionStatus, "] count elements [", len(updateTaskState.Repositories), "]")
		//for _, repo := range updateTaskState.Repositories {
		//	repositoryViewModel := &ViewModelRepository{
		//		URL:         repo.URL,
		//		Topics:      repo.Topics,
		//		Description: repo.Description,
		//	}
		//	str := fmt.Sprintf(
		//		"URL : %s\n\t\tTOPICS : [%s]\n\t\tABOUT : [%s]\n\t\tERR : [%s]",
		//		repositoryViewModel.URL,
		//		strings.Join(repositoryViewModel.Topics, ", "),
		//		repositoryViewModel.Description,
		//		repo.Err,
		//	)
		//	fmt.Println(str)
		//}
	}
	return nil
}

func (a *appService) UpdateStateTaskRepositoryIssues(updateTaskState *UpdateTaskRepositoryIssues) error {
	key := updateTaskState.ExecutionTaskStatus.TaskKey
	if value, exist := a.tasks.Get(key); !exist {
		err := errors.New("task with key [" + key + "] isn't exist ")
		runtimeinfo.LogError(err)
		return err
	} else {
		task := value.(*task)
		task.ExecutionStatus = updateTaskState.ExecutionTaskStatus.TaskCompleted
		results := task.Results.(int)
		results = results + len(updateTaskState.Issues)
		task.Results = results
		if task.ExecutionStatus {
			a.tasksChannel <- task.TaskKey
		}
		runtimeinfo.LogInfo("task with key :[", task.TaskKey, "] is competed ;[", task.ExecutionStatus, "] count elements [", len(updateTaskState.Issues), "]")
		//for _, repo := range updateTaskState.Repositories {
		//	repositoryViewModel := &ViewModelRepository{
		//		URL:         repo.URL,
		//		Topics:      repo.Topics,
		//		Description: repo.Description,
		//	}
		//	str := fmt.Sprintf(
		//		"URL : %s\n\t\tTOPICS : [%s]\n\t\tABOUT : [%s]\n\t\tERR : [%s]",
		//		repositoryViewModel.URL,
		//		strings.Join(repositoryViewModel.Topics, ", "),
		//		repositoryViewModel.Description,
		//		repo.Err,
		//	)
		//	fmt.Println(str)
		//}
	}
	return nil
}

func (a *appService) scanTasks() {
	for taskKey := range a.tasksChannel {
		taskItem, exist := a.tasks.Get(taskKey)
		if !exist {
			runtimeinfo.LogInfo("task isn't exist by key [", taskKey, "];")
			continue
		}
		task := taskItem.(*task)
		switch task.taskType {
		case TypeTaskGetRepositoriesByURL:
			runtimeinfo.LogInfo("TASK COMPLETED [", taskKey, "]")
			a.tasks.Pop(taskKey)
			break
		case TypeTaskGetRepositoryIssue:
			runtimeinfo.LogInfo("TASK COMPLETED [", taskKey, "] with count issues: [", task.Results, "]")
			a.tasks.Pop(taskKey)
			break
		}
		a.runDeferTasks(task.CollectorAddress)
	}
}

func (a *appService) runDeferTasks(collectorAddress string) {
	runtimeinfo.LogInfo("ATTEMPTING TO START DEFER TASKS. START.")
	var send = func(collectorAddress string, task *task) error {
		urlCollector := fmt.Sprintf("%s%s", collectorAddress, task.CollectorEndpoint)
		response, err := requests.POST(a.client, urlCollector, nil, task.SendBody)
		if err == nil && response.StatusCode == http.StatusOK {
			task.TaskSend = true
			task.CollectorAddress = collectorAddress
		}
		if err != nil {
			return err
		}
		if response.StatusCode != http.StatusOK {
			return errors.New("collector send status :[" + strconv.Itoa(response.StatusCode) + "]")
		}
		return nil
	}
	for taskItem := range a.tasks.IterBuffered() {
		task := taskItem.Val.(*task)
		if task.DeferSendTask && task.ExecutionStatus {
			a.tasks.Pop(task.TaskKey)
			continue
		}
		if task.DeferSendTask && !task.TaskSend {
			if err := send(collectorAddress, task); err == nil {
				runtimeinfo.LogInfo("run defer task by key :[", task.TaskKey, "]; on free collector.")
			} else {
				runtimeinfo.LogError("error in run defer task by key :[", task.TaskKey, "]; on free collector.")
				indexCollector := a.GetFreeCollector()
				if indexCollector != -1 {
					err := send(a.config.GithubCollectorsAddresses[indexCollector], task)
					if err == nil {
						task.TaskSend = true
						task.CollectorAddress = a.config.GithubCollectorsAddresses[indexCollector]
						runtimeinfo.LogInfo("run defer task by key :[", task.TaskKey, "]; on free collector.")
					}
				}
			}
		}
	}
	runtimeinfo.LogInfo("ATTEMPTING TO START DEFER TASKS. FINISH.")
}
