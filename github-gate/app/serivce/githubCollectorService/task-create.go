package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"strings"
)

const (
	gitHubApiAddress                          = "https://api.github.com"
	collectorEndpointRepositoriesDescriptions = "get/repos/by/url"
	collectorEndpointRepositoryIssues         = "get/repos/issues"
)

func (service *CollectorService) createTaskRepositoriesDescriptions(gateServiceTask itask.ITask, urls ...string) (task itask.ITask, err error) {
	var (
		taskKey          string
		sendTaskContext  *contextTaskSend
		uniqueKey        string
		repositoriesName []string
	)
	for _, url := range urls {
		name, _ := service.getRepositoryNameFromURL(url)
		repositoriesName = append(repositoriesName, name)
	}
	uniqueKey = strings.Join(repositoriesName, ",")
	if taskKey, err = service.createKeyForTask(
		RepositoriesDescription,
		gateServiceTask,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	if sendTaskContext, err = service.createSendContextForTask(
		RepositoriesDescription,
		taskKey,
		urls,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		RepositoriesDescription,
		taskKey,
		sendTaskContext,
		nil,
		gateServiceTask,
		service.eventRunTask,
		service.eventUpdateTaskDescriptionsRepositories,
	)
}

func (service *CollectorService) createTaskRepositoryIssues(gateServiceTask itask.ITask, url string) (task itask.ITask, err error) {
	var (
		taskKey         string
		sendTaskContext *contextTaskSend
		uniqueKey, _    = service.getRepositoryNameFromURL(url)
	)
	if taskKey, err = service.createKeyForTask(
		RepositoryIssues,
		gateServiceTask,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	if sendTaskContext, err = service.createSendContextForTask(
		RepositoryIssues,
		taskKey,
		url,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		RepositoryIssues,
		taskKey,
		sendTaskContext,
		nil,
		gateServiceTask,
		service.eventRunTask,
		service.eventUpdateTaskRepositoryIssues,
	)
}

func (service *CollectorService) createTriggerDescriptionRepository(gateServiceTask itask.ITask, url string) (task itask.ITask, err error) {
	var (
		taskKey         string
		sendTaskContext *contextTaskSend
		uniqueKey, _    = service.getRepositoryNameFromURL(url)
	)
	if taskKey, err = service.createKeyForTask(
		RepositoriesDescriptionAndIssues,
		gateServiceTask,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	taskKey = strings.Join([]string{"(trigger)", taskKey}, " ")
	if sendTaskContext, err = service.createSendContextForTask(
		RepositoriesDescription,
		taskKey,
		[]string{url},
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		RepositoriesDescriptionAndIssues,
		taskKey,
		sendTaskContext,
		dataModel.RepositoryModel{},
		gateServiceTask,
		service.eventRunTask,
		service.eventUpdateTriggerDescriptionRepository,
	)
}

func (service *CollectorService) createDependentRepositoryIssues(triggerTask itask.ITask, url string) (constructor itask.ITask, err error) {
	var (
		taskKey         string
		sendTaskContext *contextTaskSend
		uniqueKey, _    = service.getRepositoryNameFromURL(url)
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
		url,
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

func (service *CollectorService) createTaskRepositoriesDescriptionsAndIssues(gateServiceTask itask.ITask, urls ...string) (triggers []itask.ITask, err error) {
	var (
		countTasks = int64(len(urls)) * 2
	)
	if isFilled := service.taskManager.QueueIsFilled(countTasks); isFilled {
		return nil, gotasker.ErrorQueueIsFilled
	}
	triggers = make([]itask.ITask, 0)
	for _, url := range urls {
		trigger, err := service.createTriggerDescriptionRepository(gateServiceTask, url)
		if err != nil {
			return nil, err
		}
		dependent, err := service.createDependentRepositoryIssues(trigger, url)
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

func (service *CollectorService) createKeyForTask(taskType itask.Type, gateServiceTask itask.ITask, uniqueKey string) (taskKey string, err error) {
	var (
		gateServiceTaskKey = strings.Join(
			[]string{
				"[gate task key:{",
				gateServiceTask.GetKey(),
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
				gateServiceTaskKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	case RepositoryIssues:
		return strings.Join(
			[]string{
				"task for collector:{repository-issues-for",
				gateServiceTaskKey,
				uniqueKey,
				"}",
			}, "",
		), nil
	case RepositoriesDescriptionAndIssues:
		return strings.Join(
			[]string{
				"task for collector:{repository-description-and-issues-for",
				gateServiceTaskKey,
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
		return &contextTaskSend{
			CollectorAddress:  "",
			CollectorURL:      "",
			CollectorEndpoint: collectorEndpointForTaskContext,
			JSONBody: &jsonSendToCollectorDescriptionsRepositories{
				TaskKey: taskKey,
				URLS:    data.([]string),
			},
		}, nil
	case RepositoryIssues:
		return &contextTaskSend{
			CollectorAddress:  "",
			CollectorURL:      "",
			CollectorEndpoint: collectorEndpointForTaskContext,
			JSONBody: &jsonSendToCollectorRepositoryIssues{
				TaskKey: taskKey,
				URL:     data.(string),
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
