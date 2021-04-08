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
		RepositoriesDescription,
		taskAppService,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	if sendTaskContext, err = service.createSendContextForTask(
		RepositoriesDescription,
		taskKey,
		repositories,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		RepositoriesDescription,
		taskKey,
		sendTaskContext,
		nil,
		taskAppService,
		service.eventRunTask,
		service.eventUpdateTaskDescriptionsRepositories,
	)
}

func (service *CollectorService) createTaskRepositoryIssues(taskAppService itask.ITask, repository dataModel.RepositoryModel) (task itask.ITask, err error) {
	var (
		taskKey         string
		sendTaskContext *contextTaskSend
		uniqueKey       = repository.Name
	)
	if taskKey, err = service.createKeyForTask(
		RepositoryIssues,
		taskAppService,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	if sendTaskContext, err = service.createSendContextForTask(
		RepositoryIssues,
		taskKey,
		repository,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		RepositoryIssues,
		taskKey,
		sendTaskContext,
		nil,
		taskAppService,
		service.eventRunTask,
		service.eventUpdateTaskRepositoryIssues,
	)
}

func (service *CollectorService) createTriggerDescriptionRepository(taskAppService itask.ITask, repository dataModel.RepositoryModel) (task itask.ITask, err error) {
	var (
		taskKey         string
		sendTaskContext *contextTaskSend
		uniqueKey       = repository.Name
	)
	if taskKey, err = service.createKeyForTask(
		RepositoriesDescriptionAndIssues,
		taskAppService,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	taskKey = strings.Join([]string{"(trigger)", taskKey}, " ")
	if sendTaskContext, err = service.createSendContextForTask(
		RepositoriesDescription,
		taskKey,
		[]dataModel.RepositoryModel{repository},
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		RepositoriesDescriptionAndIssues,
		taskKey,
		sendTaskContext,
		dataModel.RepositoryModel{},
		taskAppService,
		service.eventRunTask,
		service.eventUpdateTriggerDescriptionRepository,
	)
}

func (service *CollectorService) createDependentRepositoryIssues(triggerTask itask.ITask, repository dataModel.RepositoryModel) (constructor itask.ITask, err error) {
	var (
		taskKey         string
		sendTaskContext *contextTaskSend
		uniqueKey       = repository.Name
	)
	if taskKey, err = service.createKeyForTask(
		RepositoriesDescriptionAndIssues,
		triggerTask.GetState().GetCustomFields().(itask.ITask),
		uniqueKey,
	); err != nil {
		return nil, err
	}
	taskKey = strings.Join([]string{"(dependent)", taskKey}, " ")
	if sendTaskContext, err = service.createSendContextForTask(
		RepositoryIssues,
		taskKey,
		repository,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		RepositoriesDescriptionAndIssues,
		taskKey,
		sendTaskContext,
		make([]dataModel.IssueModel, 0),
		triggerTask,
		service.eventRunTask,
		service.eventUpdateDependentRepositoryIssues,
	)
}

func (service *CollectorService) createTaskRepositoriesDescriptionsAndIssues(taskAppService itask.ITask, repositories ...dataModel.RepositoryModel) (triggers []itask.ITask, err error) {
	var (
		countTasks = int64(len(repositories)) * 2
	)
	if isFilled := service.taskManager.QueueIsFilled(countTasks); isFilled {
		return nil, gotasker.ErrorQueueIsFilled
	}
	triggers = make([]itask.ITask, 0)
	for _, repository := range repositories {
		trigger, err := service.createTriggerDescriptionRepository(taskAppService, repository)
		if err != nil {
			return nil, err
		}
		dependent, err := service.createDependentRepositoryIssues(trigger, repository)
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
	case RepositoriesDescription:
		return strings.Join(
			[]string{
				"task for collector:{repositories-descriptions-for",
				taskAppServiceKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	case RepositoryIssues:
		return strings.Join(
			[]string{
				"task for collector:{repository-issues-for",
				taskAppServiceKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	case RepositoriesDescriptionAndIssues:
		return strings.Join(
			[]string{
				"task for collector:{repository-description-and-issues-for",
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
	case RepositoriesDescription:
		var (
			jsonData = make([]jsonRepository, 0)
		)
		models := data.([]dataModel.RepositoryModel)
		for _, model := range models {
			jsonData = append(jsonData, jsonRepository{
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
	case RepositoryIssues:
		var (
			model    = data.(dataModel.RepositoryModel)
			jsonData = jsonRepository{
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
	default:
		return nil, ErrorTaskTypeNotExist
	}
}

func (service *CollectorService) getCollectorUrlForTaskContext(taskType itask.Type) (url string, err error) {
	switch taskType {
	case RepositoriesDescription:
		return collectorEndpointRepositoriesDescriptions, nil
	case RepositoryIssues:
		return collectorEndpointRepositoryIssues, nil
	default:
		return url, ErrorTaskTypeNotExist
	}
}
