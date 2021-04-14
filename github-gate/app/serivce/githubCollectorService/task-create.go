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
		TaskTypeDownloadOnlyDescriptions,
		taskAppService,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	if sendTaskContext, err = service.createSendContextForTask(
		TaskTypeDownloadOnlyDescriptions,
		taskKey,
		repositories,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		TaskTypeDownloadOnlyDescriptions,
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
		TaskTypeDownloadOnlyIssues,
		taskAppService,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	if sendTaskContext, err = service.createSendContextForTask(
		TaskTypeDownloadOnlyIssues,
		taskKey,
		repository,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		TaskTypeDownloadOnlyIssues,
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
			TaskType: TaskTypeDownloadCompositeByKeyWord,
			Fields:   dataModel.RepositoryModel{},
		})
		keywordIssues.GetState().SetEventUpdateState(service.eventUpdateDependentRepositoryAndRepositoriesKeyWord)
		keywordIssues.SetType(TaskTypeDownloadCompositeRepositoryAndRepositoriesContainingKeyWord)
		keywordDependents = append(keywordDependents, keywordIssues)
	}
	//
	keywordRepositoriesTrigger.SetType(TaskTypeDownloadCompositeRepositoryAndRepositoriesContainingKeyWord)
	keywordRepositoriesTrigger.GetState().SetCustomFields(&compositeCustomFields{
		TaskType: TaskTypeDownloadCompositeByKeyWord,
	})
	keywordRepositoriesTrigger.GetState().SetEventUpdateState(service.eventUpdateDependentRepositoryAndRepositoriesKeyWord)
	keywordRepositoriesTrigger, err = service.taskManager.ModifyTaskAsTrigger(
		keywordRepositoriesTrigger,
		keywordDependents...,
	)
	if err != nil {
		return nil, gotasker.ErrorQueueIsFilled
	}
	repositoryTrigger.SetType(TaskTypeDownloadCompositeRepositoryAndRepositoriesContainingKeyWord)
	repositoryTrigger.GetState().SetCustomFields(&compositeCustomFields{
		TaskType: TaskTypeDownloadOnlyDescriptions,
		Fields:   taskAppService,
	})
	repositoryTrigger.GetState().SetEventUpdateState(service.eventUpdateTriggerRepositoryAndRepositoriesKeyWord)
	repositoryIssues.SetType(TaskTypeDownloadCompositeRepositoryAndRepositoriesContainingKeyWord)
	repositoryIssues.GetState().SetEventUpdateState(service.eventUpdateDependentRepositoryAndRepositoriesKeyWord)
	repositoryIssues.GetState().SetCustomFields(&compositeCustomFields{
		TaskType: TaskTypeDownloadOnlyIssues,
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
	case TaskTypeDownloadOnlyDescriptions:
		return strings.Join(
			[]string{
				"task for collector:{repositories-descriptions-for",
				taskAppServiceKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	case TaskTypeDownloadOnlyIssues:
		return strings.Join(
			[]string{
				"task for collector:{repository-issues-for",
				taskAppServiceKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	case TaskTypeDownloadCompositeByName:
		return strings.Join(
			[]string{
				"task for collector:{repository-description-and-issues-for",
				taskAppServiceKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	case TaskTypeDownloadCompositeByKeyWord:
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
	case TaskTypeDownloadOnlyDescriptions:
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
	case TaskTypeDownloadOnlyIssues:
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
	case TaskTypeDownloadCompositeByKeyWord:
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
	case TaskTypeDownloadOnlyDescriptions:
		return collectorEndpointRepositoriesDescriptions, nil
	case TaskTypeDownloadOnlyIssues:
		return collectorEndpointRepositoryIssues, nil
	case TaskTypeDownloadCompositeByKeyWord:
		return collectorEndpointRepositoriesByKeyWord, nil
	default:
		return url, ErrorTaskTypeNotExist
	}
}
