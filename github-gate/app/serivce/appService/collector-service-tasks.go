package appService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"strings"
)

func (service *AppService) eventRunTaskDownloadRepositories(task itask.ITask) (doTaskAsDefer, sendToErrorChannel bool, err error) {
	switch task.GetType() {
	case ApiTaskDownloadRepositoryByName:
		err = service.collectorService.CreateTaskRepositoriesDescriptionAndIssues(
			task,
			task.GetState().GetSendContext().([]dataModel.RepositoryModel)...,
		)
		if err != nil {
			return true, false, nil
		}
		break
	}
	return false, false, nil
}

func (service *AppService) eventUpdateDownloadRepositories(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	if isTrigger, dependentsTasks := task.IsTrigger(); isTrigger {
		for next := 0; next < len(dependentsTasks); next++ {
			dependent := dependentsTasks[next]
			repositories := dependent.GetState().GetSendContext().([]dataModel.RepositoryModel)
			repositories = append(repositories, somethingUpdateContext.([]dataModel.RepositoryModel)...)
		}
	}
	task.GetState().SetCompleted(true)
	return nil, false
}



func (service *AppService) gettingResultFromCollectorService() {
	for task := range service.channelResultsFromCollector {
		repositories := task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		service.taskManager.SetUpdateForTask(
			task.GetKey(),
			repositories,
		)
	}
}

func (service *AppService) createTaskDownloadRepositoriesByName(taskType itask.Type, jsonModel *ApiJsonDownloadRepositoriesByName) (task itask.ITask, err error) {
	var (
		taskKey           string
		repositoriesNames = make([]string, 0)
		sendContext       = make([]dataModel.RepositoryModel, 0)
		updateContext     = make([]dataModel.RepositoryModel, 0)
		customFields      = service.channelResultsFromCollector
	)
	for _, repository := range jsonModel.Repositories {
		if strings.TrimSpace(repository.Name) == "" || strings.TrimSpace(repository.Owner) == "" {
			return nil, ErrorEmptyOrIncompleteJSONData
		}
		sendContext = append(sendContext, dataModel.RepositoryModel{
			Name:  repository.Name,
			Owner: repository.Owner,
		})
		repositoriesNames = append(repositoriesNames, repository.Name)
	}
	taskKey = strings.Join([]string{
		"download-repositories-by-name",
		"{",
		strings.Join(repositoriesNames, "-"),
		"}",
	}, "")
	return service.taskManager.CreateTask(
		taskType,
		taskKey,
		sendContext,
		updateContext,
		customFields,
		service.eventRunTaskDownloadRepositories,
		service.eventUpdateDownloadRepositories,
	)
}
