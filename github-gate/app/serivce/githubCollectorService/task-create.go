package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"strings"
)

func (service *CollectorService) createTaskDownloadRepositoriesDescription(
	taskType itask.Type,
	repositoryDataModels dataModel.RepositoryModel,
) (task itask.ITask, err error) {
	var (
		collectorEndpoint, taskKey string
		sendContext                *contextTaskSend
		//
		uniqueKey     = repositoryDataModels.Name
		updateContext = new(dataModel.RepositoryModel)
		jsonBody      = []jsonSendToCollectorRepository{
			{
				Name:  repositoryDataModels.Name,
				Owner: repositoryDataModels.Owner,
			},
		}
	)
	collectorEndpoint = collectorEndpointRepositoriesDescriptions
	taskKey = strings.Join(
		[]string{
			"(download-descriptions){",
			uniqueKey,
			"}",
		},
		"",
	)
	sendContext = &contextTaskSend{
		CollectorAddress:  "",
		CollectorURL:      "",
		CollectorEndpoint: collectorEndpoint,
		JSONBody: &jsonSendToCollectorDescriptionsRepositories{
			TaskKey:      taskKey,
			Repositories: jsonBody,
		},
	}
	return service.taskManager.CreateTask(
		taskType,
		taskKey,
		sendContext,
		updateContext,
		nil,
		nil,
		nil,
	)
}

func (service *CollectorService) createTaskDownloadRepositoryIssues(
	taskType itask.Type,
	repositoryDataModel dataModel.RepositoryModel,
) (task itask.ITask, err error) {
	var (
		collectorEndpoint, taskKey string
		sendContext                *contextTaskSend
		updateContext              interface{}
		//
		uniqueKey = repositoryDataModel.Name
		jsonBody  = jsonSendToCollectorRepository{
			Name:  repositoryDataModel.Name,
			Owner: repositoryDataModel.Owner,
		}
	)
	updateContext = nil
	collectorEndpoint = collectorEndpointRepositoryIssues
	taskKey = strings.Join(
		[]string{
			"(download-issues){",
			uniqueKey,
			"}",
		},
		"",
	)
	sendContext = &contextTaskSend{
		CollectorAddress:  "",
		CollectorURL:      "",
		CollectorEndpoint: collectorEndpoint,
		JSONBody: &jsonSendToCollectorRepositoryIssues{
			TaskKey:    taskKey,
			Repository: jsonBody,
		},
	}
	return service.taskManager.CreateTask(
		taskType,
		taskKey,
		sendContext,
		updateContext,
		nil,
		nil,
		nil,
	)
}

func (service *CollectorService) createTaskDownloadRepositoriesByKeyword(
	taskType itask.Type,
	keyWord string,
) (task itask.ITask, err error) {
	var (
		collectorEndpoint, taskKey string
		sendContext                *contextTaskSend
		//
		updateContext = make([]dataModel.RepositoryModel, 0)
		uniqueKey     = keyWord
		jsonBody      = jsonSendToCollectorRepositoriesByKeyWord{}
	)
	collectorEndpoint = collectorEndpointRepositoriesByKeyWord
	taskKey = strings.Join(
		[]string{
			"(download-repositories-keyword){",
			uniqueKey,
			"}",
		},
		"",
	)
	jsonBody.TaskKey = taskKey
	jsonBody.KeyWord = keyWord
	sendContext = &contextTaskSend{
		CollectorAddress:  "",
		CollectorURL:      "",
		CollectorEndpoint: collectorEndpoint,
		JSONBody:          jsonBody,
	}
	return service.taskManager.CreateTask(
		taskType,
		taskKey,
		sendContext,
		updateContext,
		nil,
		nil,
		nil,
	)
}
