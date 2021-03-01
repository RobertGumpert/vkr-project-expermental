package serivce

import (
	"errors"
	"fmt"
	"github-gate/app/models"
	"github-gate/pckg/requests"
	"github-gate/pckg/runtimeinfo"
	"net/http"
	"strings"
)

func (a *AppService) CreateTaskRepositoriesByURL(urls []string, deferTask bool) error {
	var (
		taskKey           = ""
		collectorEndpoint = "/get/repos/by/url"
		sendBody          = &models.SendTaskReposByURL{
			TaskKey: taskKey,
			URLS:    urls,
		}
	)
	if deferTask {
		taskKey = fmt.Sprintf("defer task key [%d]", a.tasks.Count()+1)
	} else {
		taskKey = fmt.Sprintf("key [%d]", a.tasks.Count()+1)
	}
	sendBody.TaskKey = taskKey
	repositories := make([]*models.ViewModelRepository, 0)
	task := &task{
		taskType:        TypeTaskGetRepositoriesByURL,
		TaskKey:         taskKey,
		ExecutionStatus: false,
		About: strings.Join(
			[]string{
				"get repositories by url: [",
				strings.Join(urls, ", "),
				"]",
			}, " "),
		CollectorAddress:  "",
		Results:           repositories,
		DeferSendTask:     deferTask,
		TaskSend:          false,
		SendBody:          sendBody,
		CollectorEndpoint: collectorEndpoint,
	}
	if deferTask {
		a.sendDeferTask(task)
		return nil
	}
	return a.sendNonDeferTask(task)
}

func (a *AppService) CreateTaskGetRepositoriesIssues(urls []string, deferTask bool) error {
	for _, url := range urls {
		var (
			taskKey           = ""
			collectorEndpoint = "/get/repos/issues"
			sendBody          = &models.SendTaskRepositoryIssues{
				TaskKey: taskKey,
				URL:     url,
			}
		)
		if deferTask {
			taskKey = fmt.Sprintf("defer task key [%d]", a.tasks.Count()+1)
		} else {
			taskKey = fmt.Sprintf("key [%d]", a.tasks.Count()+1)
		}
		sendBody.TaskKey = taskKey
		task := &task{
			taskType:        TypeTaskGetRepositoryIssue,
			TaskKey:         taskKey,
			ExecutionStatus: false,
			About: strings.Join(
				[]string{
					"get issues for repositories: [",
					strings.Join(urls, ", "),
					"]",
				}, " "),
			CollectorAddress:  "",
			Results:           0,
			DeferSendTask:     deferTask,
			TaskSend:          false,
			SendBody:          sendBody,
			CollectorEndpoint: collectorEndpoint,
		}
		if deferTask {
			a.sendDeferTask(task)
			return nil
		} else {
			return a.sendNonDeferTask(task)
		}
	}
	return nil
}

func (a *AppService) CreateTaskGetRepositoriesAndIssues(urls []string, deferTask bool) error {
	err := a.CreateTaskRepositoriesByURL(urls, deferTask)
	return err
}

func (a *AppService) sendDeferTask(task *task) {
	a.tasks.Set(task.TaskKey, task)
	response, err := a.sendTask(task)
	if err != nil {
		runtimeinfo.LogInfo("non send defer task by key :[", task.TaskKey, "];")
		return
	}
	if response.StatusCode != http.StatusOK {
		runtimeinfo.LogInfo("non send defer task by key :[", task.TaskKey, "];")
		return
	}
	runtimeinfo.LogInfo("run defer task by key :[", task.TaskKey, "]; on free collector.")
}

func (a *AppService) sendNonDeferTask(task *task) error {
	if a.tasks.Count() > int(a.config.CountTask/2) {
		return errors.New("No place in the queue. ")
	}
	response, err := a.sendTask(task)
	if err != nil {
		return err
	} else {
		a.tasks.Set(task.TaskKey, task)
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("send task on collector finish with non 200 status")
	} else {
		a.tasks.Set(task.TaskKey, task)
	}
	return nil
}

func (a *AppService) sendTask(task *task) (*http.Response, error) {
	indexCollector := a.GetFreeCollector()
	if indexCollector != -1 {
		urlCollector := fmt.Sprintf("%s%s", a.config.GithubCollectorsAddresses[indexCollector], task.CollectorEndpoint)
		response, err := requests.POST(a.client, urlCollector, nil, task.SendBody)
		if err == nil && response.StatusCode == http.StatusOK {
			task.TaskSend = true
			task.CollectorAddress = a.config.GithubCollectorsAddresses[indexCollector]
		}
		return response, err
	}
	return nil, errors.New("Non free GitHub collectors. ")
}
