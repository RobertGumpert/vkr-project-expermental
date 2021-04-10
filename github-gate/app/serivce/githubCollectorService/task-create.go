package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"strings"
)

func (service *CollectorService) createTaskOnlyRepositoriesDescriptions(
	taskAppService itask.ITask,
	repositories ...dataModel.RepositoryModel,
) (task itask.ITask, err error) {
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
		OnlyDescriptions,
		taskAppService,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	if sendTaskContext, err = service.createSendContextForTask(
		OnlyDescriptions,
		taskKey,
		repositories,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		OnlyDescriptions,
		taskKey,
		sendTaskContext,
		nil,
		taskAppService,
		service.eventRunTask,
		service.eventUpdateOnlyDescriptions,
	)
}

func (service *CollectorService) createTaskOnlyRepositoryIssues(
	taskAppService itask.ITask,
	repository dataModel.RepositoryModel,
) (task itask.ITask, err error) {
	var (
		taskKey         string
		sendTaskContext *contextTaskSend
		uniqueKey       = repository.Name
	)
	if taskKey, err = service.createKeyForTask(
		OnlyIssues,
		taskAppService,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	if sendTaskContext, err = service.createSendContextForTask(
		OnlyIssues,
		taskKey,
		repository,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		OnlyIssues,
		taskKey,
		sendTaskContext,
		nil,
		taskAppService,
		service.eventRunTask,
		service.eventUpdateOnlyIssues,
	)
}

func (service *CollectorService) createCompositeTaskSearchByName(
	taskAppService itask.ITask,
	repositories ...dataModel.RepositoryModel,
) (triggers []itask.ITask, err error) {
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

func (service *CollectorService) createCompositeTaskSearchByKeyWord(
	taskAppService itask.ITask,
	keyWord string,
) (trigger itask.ITask, err error) {
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

func (service *CollectorService) createTaskRepositoryAndRepositoriesContainingKeyWord(
	taskAppService itask.ITask,
	repository dataModel.RepositoryModel,
	keyWord string,
) (task itask.ITask, err error) {
	var (
		countTasks           = int64(33)
		repositoryDependents = make([]itask.ITask, 0)
		keywordDependents    = make([]itask.ITask, 0)
	)
	if isFilled := service.taskManager.QueueIsFilled(countTasks); isFilled {
		return nil, gotasker.ErrorQueueIsFilled
	}
	repositoryTrigger, err := service.createTriggerRepositoryByName(taskAppService, repository)
	if err != nil {
		return nil, err
	}
	repositoryIssues, err := service.createDependentIssuesByName(repositoryTrigger, repository)
	if err != nil {
		return nil, err
	}
	keywordRepositoriesTrigger, err := service.createTriggerRepositoriesByKeyWord(
		taskAppService,
		keyWord,
	)
	if err != nil {
		return nil, err
	}
	for next := 0; next < 30; next++ {
		keywordIssues, err := service.createDependentIssuesByKeyWord(keywordRepositoriesTrigger, next, keyWord)
		if err != nil {
			return nil, err
		}
		keywordIssues.GetState().SetCustomFields(&compositeCustomFields{
			TaskType: CompositeByKeyWord,
			Fields:   dataModel.RepositoryModel{},
		})
		keywordIssues.GetState().SetEventUpdateState(service.eventUpdateDependentRepositoryAndRepositoriesKeyWord)
		keywordIssues.SetType(RepositoryAndRepositoriesContainingKeyWord)
		keywordDependents = append(keywordDependents, keywordIssues)
	}
	//
	keywordRepositoriesTrigger.SetType(RepositoryAndRepositoriesContainingKeyWord)
	keywordRepositoriesTrigger.GetState().SetCustomFields(&compositeCustomFields{
		TaskType: CompositeByKeyWord,
	})
	keywordRepositoriesTrigger.GetState().SetEventUpdateState(service.eventUpdateDependentRepositoryAndRepositoriesKeyWord)
	keywordRepositoriesTrigger, err = service.taskManager.ModifyTaskAsTrigger(
		keywordRepositoriesTrigger,
		keywordDependents...,
	)
	if err != nil {
		return nil, gotasker.ErrorQueueIsFilled
	}
	repositoryTrigger.SetType(RepositoryAndRepositoriesContainingKeyWord)
	repositoryTrigger.GetState().SetCustomFields(&compositeCustomFields{
		TaskType: OnlyDescriptions,
		Fields:   taskAppService,
	})
	repositoryTrigger.GetState().SetEventUpdateState(service.eventUpdateTriggerRepositoryAndRepositoriesKeyWord)
	repositoryIssues.SetType(RepositoryAndRepositoriesContainingKeyWord)
	repositoryIssues.GetState().SetEventUpdateState(service.eventUpdateDependentRepositoryAndRepositoriesKeyWord)
	repositoryIssues.GetState().SetCustomFields(&compositeCustomFields{
		TaskType: OnlyIssues,
		Fields:   dataModel.RepositoryModel{},
	})
	repositoryDependents = append(repositoryDependents, repositoryIssues, keywordRepositoriesTrigger)
	service.taskManager.SetRunBan(repositoryDependents...)
	return service.taskManager.ModifyTaskAsTrigger(
		repositoryTrigger,
		repositoryDependents...,
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
	case OnlyDescriptions:
		return strings.Join(
			[]string{
				"task for collector:{repositories-descriptions-for",
				taskAppServiceKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	case OnlyIssues:
		return strings.Join(
			[]string{
				"task for collector:{repository-issues-for",
				taskAppServiceKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	case CompositeByName:
		return strings.Join(
			[]string{
				"task for collector:{repository-description-and-issues-for",
				taskAppServiceKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	case CompositeByKeyWord:
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
	case OnlyDescriptions:
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
	case OnlyIssues:
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
	case CompositeByKeyWord:
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
	case OnlyDescriptions:
		return collectorEndpointRepositoriesDescriptions, nil
	case OnlyIssues:
		return collectorEndpointRepositoryIssues, nil
	case CompositeByKeyWord:
		return collectorEndpointRepositoriesByKeyWord, nil
	default:
		return url, ErrorTaskTypeNotExist
	}
}
