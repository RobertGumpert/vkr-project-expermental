package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"strings"
)

func (service *CollectorService) createTaskRepositoriesDescriptions(taskAppService itask.ITask, repositories ...dataModel.RepositoryModel) (task itask.ITask, err error) {
	var (
		taskKey          string
		sendTaskContext  *contextTaskSend
		uniqueKey        string
		repositoriesName []string
	)
	for _, repository := range repositories {
		repositoriesName = append(repositoriesName, repository.Name)
	}
	uniqueKey = strings.Join(repositoriesName, ",")
	if taskKey, err = service.createKeyForTask(
		RepositoriesOnlyDescription,
		taskAppService,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	if sendTaskContext, err = service.createSendContextForTask(
		RepositoriesOnlyDescription,
		taskKey,
		repositories,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		RepositoriesOnlyDescription,
		taskKey,
		sendTaskContext,
		nil,
		taskAppService,
		service.eventRunTask,
		service.eventUpdateSingleDescriptionsRepositories,
	)
}

func (service *CollectorService) createTaskRepositoryIssues(taskAppService itask.ITask, repository dataModel.RepositoryModel) (task itask.ITask, err error) {
	var (
		taskKey         string
		sendTaskContext *contextTaskSend
		uniqueKey       = repository.Name
	)
	if taskKey, err = service.createKeyForTask(
		RepositoryOnlyIssues,
		taskAppService,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	if sendTaskContext, err = service.createSendContextForTask(
		RepositoryOnlyIssues,
		taskKey,
		repository,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		RepositoryOnlyIssues,
		taskKey,
		sendTaskContext,
		nil,
		taskAppService,
		service.eventRunTask,
		service.eventUpdateSingleRepositoryIssues,
	)
}

func (service *CollectorService) createTaskRepositoriesDescriptionsAndIssuesByName(taskAppService itask.ITask, repositories ...dataModel.RepositoryModel) (triggers []itask.ITask, err error) {
	var (
		countTasks = int64(len(repositories)) * 2
	)
	if isFilled := service.taskManager.QueueIsFilled(countTasks); isFilled {
		return nil, gotasker.ErrorQueueIsFilled
	}
	triggers = make([]itask.ITask, 0)
	for _, repository := range repositories {
		trigger, err := service.createTriggerRepositoryByName(taskAppService, repository)
		if err != nil {
			return nil, err
		}
		dependent, err := service.createDependentIssuesByName(trigger, repository)
		if err != nil {
			return nil, err
		}
		trigger, err = service.taskManager.ModifyTaskAsTrigger(
			trigger,
			dependent,
		)
		triggers = append(triggers, trigger)
		if err != nil {
			return nil, err
		}
	}
	return triggers, nil
}

func (service *CollectorService) createTaskRepositoriesDescriptionsAndIssuesByKeyWord(taskAppService itask.ITask, keyWord string) (trigger itask.ITask, err error) {
	var (
		countTasks      = int64(31)
		dependentsTasks = make([]itask.ITask, 0)
	)
	if isFilled := service.taskManager.QueueIsFilled(countTasks); isFilled {
		return nil, gotasker.ErrorQueueIsFilled
	}
	trigger, err = service.createTriggerRepositoriesByKeyWord(
		taskAppService,
		keyWord,
	)
	if err != nil {
		return nil, err
	}
	for next := 0; next < int(countTasks-1); next++ {
		dependent, err := service.createDependentIssuesByKeyWord(trigger, next, keyWord)
		if err != nil {
			return nil, err
		}
		dependentsTasks = append(dependentsTasks, dependent)
	}
	return service.taskManager.ModifyTaskAsTrigger(
		trigger,
		dependentsTasks...,
	)
}

func (service *CollectorService) createKeyForTask(taskType itask.Type, taskAppService itask.ITask, uniqueKey string) (taskKey string, err error) {
	var (
		taskAppServiceKey = strings.Join(
			[]string{
				"[gate task key:{",
				taskAppService.GetKey(),
				"}]",
			},
			"",
		)
	)
	uniqueKey = strings.Join(
		[]string{
			"[unique:{",
			uniqueKey,
			"}]",
		},
		"",
	)
	switch taskType {
	case RepositoriesOnlyDescription:
		return strings.Join(
			[]string{
				"task for collector:{repositories-descriptions-for",
				taskAppServiceKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	case RepositoryOnlyIssues:
		return strings.Join(
			[]string{
				"task for collector:{repository-issues-for",
				taskAppServiceKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	case RepositoryByName:
		return strings.Join(
			[]string{
				"task for collector:{repository-description-and-issues-for",
				taskAppServiceKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	case RepositoriesByKeyWord:
		return strings.Join(
			[]string{
				"task for collector:{repositories-by-keyword",
				taskAppServiceKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	default:
		return taskKey, ErrorTaskTypeNotExist
	}
}

func (service *CollectorService) createSendContextForTask(taskType itask.Type, taskKey string, data interface{}) (context *contextTaskSend, err error) {
	var (
		collectorEndpointForTaskContext string
	)
	collectorEndpointForTaskContext, err = service.getCollectorUrlForTaskContext(taskType)
	if err != nil {
		return nil, err
	}
	switch taskType {
	case RepositoriesOnlyDescription:
		var (
			jsonData = make([]jsonSendToCollectorRepository, 0)
		)
		models := data.([]dataModel.RepositoryModel)
		for _, model := range models {
			jsonData = append(jsonData, jsonSendToCollectorRepository{
				Name:  model.Name,
				Owner: model.Owner,
			})
		}
		return &contextTaskSend{
			CollectorAddress:  "",
			CollectorURL:      "",
			CollectorEndpoint: collectorEndpointForTaskContext,
			JSONBody: &jsonSendToCollectorDescriptionsRepositories{
				TaskKey:      taskKey,
				Repositories: jsonData,
			},
		}, nil
	case RepositoryOnlyIssues:
		var (
			model    = data.(dataModel.RepositoryModel)
			jsonData = jsonSendToCollectorRepository{
				Name:  model.Name,
				Owner: model.Owner,
			}
		)
		return &contextTaskSend{
			CollectorAddress:  "",
			CollectorURL:      "",
			CollectorEndpoint: collectorEndpointForTaskContext,
			JSONBody: &jsonSendToCollectorRepositoryIssues{
				TaskKey:    taskKey,
				Repository: jsonData,
			},
		}, nil
	case RepositoriesByKeyWord:
		var (
			jsonData = data.(string)
		)
		return &contextTaskSend{
			CollectorAddress:  "",
			CollectorURL:      "",
			CollectorEndpoint: collectorEndpointForTaskContext,
			JSONBody: &jsonSendToCollectorRepositoriesByKeyWord{
				TaskKey: taskKey,
				KeyWord: jsonData,
			},
		}, nil
	default:
		return nil, ErrorTaskTypeNotExist
	}
}

func (service *CollectorService) getCollectorUrlForTaskContext(taskType itask.Type) (url string, err error) {
	switch taskType {
	case RepositoriesOnlyDescription:
		return collectorEndpointRepositoriesDescriptions, nil
	case RepositoryOnlyIssues:
		return collectorEndpointRepositoryIssues, nil
	case RepositoriesByKeyWord:
		return collectorEndpointRepositoriesByKeyWord, nil
	default:
		return url, ErrorTaskTypeNotExist
	}
}
