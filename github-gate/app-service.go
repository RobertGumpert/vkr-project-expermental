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
	TypeTaskGetReposByURL taskType = 0
)

type task struct {
	taskType        taskType
	ExecutionStatus bool
	About           string
	TaskKey         string
	Collector       string
	Results         interface{}
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

func (a *appService) CreateTaskReposByURL(createTaskModel *CreateTaskRepoByURLS) error {
	collector := a.GetFreeCollector()
	if collector == -1 {
		err := errors.New("Non free GitHub collectors. ")
		runtimeinfo.LogError(err)
		return err
	}
	repositories := make([]*ViewModelRepository, 0)
	task := &task{
		taskType:        TypeTaskGetReposByURL,
		ExecutionStatus: false,
		About:           strings.Join(createTaskModel.Repositories, ", "),
		TaskKey:         fmt.Sprintf("key [%d]", a.tasks.Count()+1),
		Collector:       a.config.GithubCollectorsAddresses[collector],
		Results:         &repositories,
	}
	a.tasks.Set(task.TaskKey, task)
	url := fmt.Sprintf("%s%s", task.Collector, "/get/repos/by/url")
	response, err := requests.POST(a.client, url, nil, &SendTaskReposByURL{
		TaskKey: task.TaskKey,
		URLS:    createTaskModel.Repositories,
	})
	if err != nil {
		runtimeinfo.LogError("[", url, "] error: ", err)
		return err
	}
	status := "[" + url + "] status: " + strconv.Itoa(response.StatusCode)
	if response.StatusCode == http.StatusOK {
		runtimeinfo.LogInfo(status)
	} else {
		err := errors.New(status)
		runtimeinfo.LogError(err)
		return err
	}
	return nil
}

func (a *appService) UpdateStateTaskReposByURL(updateTaskState *UpdateTaskReposByURLS) error {
	key := updateTaskState.ExecutionTaskStatus.TaskKey
	if value, exist := a.tasks.Get(key); !exist {
		err := errors.New("task with key [" + key + "] isn't exist ")
		runtimeinfo.LogError(err)
		return err
	} else {
		task := value.(*task)
		if task.ExecutionStatus {
			a.tasksChannel <- task.TaskKey
		}
		for _, repo := range updateTaskState.Repositories {
			repositoryViewModel := &ViewModelRepository{
				URL:         repo.URL,
				Topics:      repo.Topics,
				Description: repo.Description,
			}
			str := fmt.Sprintf(
				"URL : %s\n\t\tTOPICS : [%s]\n\t\tABOUT : [%s]\n\t\tERR : [%s]",
				repositoryViewModel.URL,
				strings.Join(repositoryViewModel.Topics, ", "),
				repositoryViewModel.Description,
				repo.Err,
			)
			fmt.Println(str)
		}
	}
	return nil
}

func (a *appService) scanTasks() {
	for taskKey := range a.tasksChannel {
		if taskItem, exist := a.tasks.Get(taskKey); !exist {
			runtimeinfo.LogInfo("task isn't exist by key [", taskKey, "];")
			continue
		} else {
			task := taskItem.(*task)
			switch task.taskType {
			case TypeTaskGetReposByURL:
				runtimeinfo.LogInfo("TASK COMPLETED [", taskKey, "]")
				a.tasks.Pop(taskKey)
				break
			}
		}
	}

	//for {
	//	runtime.Gosched()
	//	for item := range a.tasks.IterBuffered() {
	//		task := item.Val.(*task)
	//		if task.ExecutionStatus == true {
	//			switch task.taskType {
	//			case TypeTaskGetReposByURL:
	//				current := task.Results.(*[]*ViewModelRepository)
	//				for i := 0; i < len(*current); i++ {
	//					str := fmt.Sprintf(
	//						"URL : %s\n\t\tTOPICS : [%s]\n\t\tABOUT : [%s]",
	//						(*current)[i].URL,
	//						strings.Join((*current)[i].Topics, ", "),
	//						(*current)[i].About,
	//					)
	//					fmt.Println(str)
	//				}
	//				runtimeinfo.LogInfo("TASK COMPLETED [", item.Key, "]")
	//				a.tasks.Pop(item.Key)
	//				break
	//			}
	//		}
	//	}
	//}
}
