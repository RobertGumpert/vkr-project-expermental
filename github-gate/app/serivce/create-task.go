package serivce

import (
	"errors"
	"fmt"
	"github-gate/app/models/dataModel"
	"github-gate/app/models/interapplicationModels/githubCollectorModels"
	"github-gate/pckg/requests"
	"github-gate/pckg/runtimeinfo"
	"net/http"
	"strings"
)

func (a *AppService) CreateTaskRepositoriesByURL(urls []string, deferTask bool, relatedTasks []string, typeTask ...taskType) error {
	if err := a.possibleAddTasks(1); err != nil {
		return err
	}
	taskTypeSelected := TypeTaskGetRepositoriesByURL
	if len(typeTask) != 0 {
		taskTypeSelected = typeTask[0]
	}
	var (
		taskKey           = ""
		collectorEndpoint = "/get/repos/by/url"
		sendBody          = &githubCollectorModels.SendTaskRepositoriesByURLS{
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
	repositories := make([]*dataModel.Repository, 0)
	task := &task{
		taskType:        taskTypeSelected,
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
		SignalTaskSend:    false,
		RelatedTasks:      relatedTasks,
		TaskSend:          false,
		SendBody:          sendBody,
		CollectorEndpoint: collectorEndpoint,
	}
	if deferTask {
		return a.sendDeferTask(task)
	}
	return a.sendNonDeferTask(task)
}

func (a *AppService) CreateTaskGetRepositoriesIssues(urls []string, deferTask, signalTaskSend bool) (error, []string, []string) {
	var relatedTasks []string
	if err := a.possibleAddTasks(len(urls)); err != nil {
		return err, nil, nil
	}
	if signalTaskSend {
		relatedTasks = make([]string, 0)
	}
	for index, url := range urls {
		var (
			taskKey           = ""
			collectorEndpoint = "/get/repos/issues"
			sendBody          = &githubCollectorModels.SendTaskRepositoryIssues{
				TaskKey: taskKey,
				URL:     url,
			}
		)
		if deferTask {
			taskKey = fmt.Sprintf("defer task key [%d]", a.tasks.Count()+1)
		} else {
			taskKey = fmt.Sprintf("key [%d]", a.tasks.Count()+1)
		}
		if signalTaskSend {
			taskKey = fmt.Sprintf("signal task key [%d]", a.tasks.Count()+1)
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
			SignalTaskSend:    signalTaskSend,
			RelatedTasks:      nil,
			TaskSend:          false,
			SendBody:          sendBody,
			CollectorEndpoint: collectorEndpoint,
		}
		if signalTaskSend {
			task.DeferSendTask = true
			a.tasks.Set(task.TaskKey, task)
			relatedTasks = append(relatedTasks, task.TaskKey)
			continue
		}
		if deferTask {
			err := a.sendDeferTask(task)
			if err != nil {
				return nil, urls[index:], nil
			}
			continue
		} else {
			err := a.sendNonDeferTask(task)
			if err != nil {
				return nil, urls[index:], nil
			}
			continue
		}
	}
	return nil, nil, relatedTasks
}

func (a *AppService) CreateTaskGetRepositoriesAndIssues(urls []string) error {
	_, _, relatedTasks := a.CreateTaskGetRepositoriesIssues(
		urls,
		true,
		true,
	)
	err := a.CreateTaskRepositoriesByURL(urls, true, relatedTasks, TypeTaskGetRepositoriesAndIssues)
	if err != nil {
		for _, task := range relatedTasks {
			a.tasks.Pop(task)
		}
		return err
	}
	return err
}

func (a *AppService) possibleAddTasks(count int) error {
	if a.tasks.Count()+count > int(a.config.SizeQueueTasksForGithubCollectors) {
		return errors.New("No place in the queue. ")
	}
	return nil
}

func (a *AppService) sendDeferTask(task *task) error {
	if err := a.possibleAddTasks(1); err != nil {
		return err
	}
	a.tasks.Set(task.TaskKey, task)
	response, err := a.sendTask(task)
	if err != nil {
		runtimeinfo.LogInfo("non send defer task by key :[", task.TaskKey, "];")
		return nil
	}
	if response.StatusCode != http.StatusOK {
		runtimeinfo.LogInfo("non send defer task by key :[", task.TaskKey, "];")
		return nil
	}
	runtimeinfo.LogInfo("run defer task by key :[", task.TaskKey, "]; on free collector.")
	return nil
}

func (a *AppService) sendNonDeferTask(task *task) error {
	if a.tasks.Count() > int(a.config.SizeQueueTasksForGithubCollectors/2) {
		return errors.New("No place in the queue. ")
	}
	response, err := a.sendTask(task)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("send task on collector finish with non 200 status")
	}
	a.tasks.Set(task.TaskKey, task)
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
