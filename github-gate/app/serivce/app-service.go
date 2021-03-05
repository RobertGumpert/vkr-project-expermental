package serivce

import (
	"errors"
	"fmt"
	"github-gate/app/config"
	"github-gate/pckg/requests"
	"github-gate/pckg/runtimeinfo"
	cmap "github.com/streamrail/concurrent-map"
	"net/http"
	"strconv"
)

type taskType uint

const (
	TypeTaskGetRepositoriesByURL     taskType = 0
	TypeTaskGetRepositoryIssue       taskType = 1
	TypeTaskGetRepositoriesAndIssues taskType = 2
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
	SignalTaskSend    bool
	RelatedTasks      []string
	TaskSend          bool
	SendBody          interface{}
	CollectorEndpoint string
	//
}

type AppService struct {
	config       *config.Config
	client       *http.Client
	tasks        *cmap.ConcurrentMap
	tasksChannel chan string
}

func NewAppService(config *config.Config) *AppService {
	tasks := cmap.New()
	a := &AppService{
		config:       config,
		client:       new(http.Client),
		tasks:        &tasks,
		tasksChannel: make(chan string, config.CountTask),
	}
	go a.scanTasks()
	return a
}

func (a *AppService) GetFreeCollector() int {
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

func (a *AppService) scanTasks() {
	for taskKey := range a.tasksChannel {
		taskItem, exist := a.tasks.Get(taskKey)
		if !exist {
			runtimeinfo.LogInfo("SCAN TASKS: task isn't exist by key [", taskKey, "];")
			continue
		}
		currentTask := taskItem.(*task)
		switch currentTask.taskType {
		case TypeTaskGetRepositoriesByURL:
			runtimeinfo.LogInfo("SCAN TASKS: TASK COMPLETED [", taskKey, "]")
			a.tasks.Pop(taskKey)
			break
		case TypeTaskGetRepositoryIssue:
			runtimeinfo.LogInfo("SCAN TASKS: TASK COMPLETED [", taskKey, "] with count issues: [", currentTask.Results, "]")
			a.tasks.Pop(taskKey)
			break
		case TypeTaskGetRepositoriesAndIssues:
			runtimeinfo.LogInfo("SCAN TASKS: TASK COMPLETED [", taskKey, "]")
			for _, relatedTaskKey := range currentTask.RelatedTasks {
				if relatedTaskValue, exist := a.tasks.Get(relatedTaskKey); exist {
					runtimeinfo.LogInfo("SCAN TASKS: update state defer task [", relatedTaskKey, "] by signal from task [",taskKey,"]")
					relatedTask := relatedTaskValue.(*task)
					relatedTask.SignalTaskSend = false
				}
			}
			a.tasks.Pop(taskKey)
			break
		}
		a.runDeferTasks(currentTask.CollectorAddress)
	}
}

func (a *AppService) runDeferTasks(collectorAddress string) {
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
		if task.DeferSendTask && task.TaskSend && task.ExecutionStatus {
			a.tasks.Pop(task.TaskKey)
			continue
		}
		if task.DeferSendTask && !task.TaskSend && !task.SignalTaskSend {
			if err := send(collectorAddress, task); err == nil {
				runtimeinfo.LogInfo("RUN DEFER TASK: run task [", task.TaskKey, "] on the vacated.")
			} else {
				runtimeinfo.LogError("RUN DEFER TASK: run defer task [", task.TaskKey, "] on the vacated ended error.")
				indexCollector := a.GetFreeCollector()
				if indexCollector != -1 {
					err := send(a.config.GithubCollectorsAddresses[indexCollector], task)
					if err == nil {
						task.TaskSend = true
						task.CollectorAddress = a.config.GithubCollectorsAddresses[indexCollector]
						runtimeinfo.LogInfo("RUN DEFER TASK: run defer task [", task.TaskKey, "] on free collector.")
					}
				}
			}
		}
	}
	runtimeinfo.LogInfo("ATTEMPTING TO START DEFER TASKS. FINISH.")
}
