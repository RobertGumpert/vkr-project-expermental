package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"strconv"
	"strings"
)

func (service *CollectorService) createTriggerRepositoriesByKeyWord(taskAppService itask.ITask, keyWord string) (task itask.ITask, err error) {
	var (
		taskKey           string
		customFields      = taskAppService
		updateTaskContext = make([]dataModel.RepositoryModel, 0)
		sendTaskContext   *contextTaskSend
		uniqueKey         = keyWord
	)
	if taskKey, err = service.createKeyForTask(
		CompositeByKeyWord,
		taskAppService,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	taskKey = strings.Join([]string{"(trigger)", taskKey}, " ")
	if sendTaskContext, err = service.createSendContextForTask(
		CompositeByKeyWord,
		taskKey,
		keyWord,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		CompositeByKeyWord,
		taskKey,
		sendTaskContext,
		updateTaskContext,
		customFields,
		service.eventRunTask,
		service.eventUpdateTriggerByKeyWord,
	)
}

func (service *CollectorService) createDependentIssuesByKeyWord(triggerTask itask.ITask, number int, keyWord string) (constructor itask.ITask, err error) {
	var (
		taskKey         string
		sendTaskContext *contextTaskSend
		customFields    = &compositeCustomFields{
			Fields: dataModel.RepositoryModel{},
		}
		updateTaskContext = make([]dataModel.IssueModel, 0)
		uniqueKey         = "%s"
	)
	if taskKey, err = service.createKeyForTask(
		CompositeByKeyWord,
		triggerTask.GetState().GetCustomFields().(itask.ITask),
		uniqueKey,
	); err != nil {
		return nil, err
	}
	dependentNumberKey := strings.Join([]string{"(dependent-", strconv.Itoa(number), ")"}, "")
	taskKey = strings.Join([]string{dependentNumberKey, taskKey}, " ")
	if sendTaskContext, err = service.createSendContextForTask(
		OnlyIssues,
		taskKey,
		dataModel.RepositoryModel{
			Name:  keyWord,
			Owner: keyWord,
		},
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		CompositeByKeyWord,
		taskKey,
		sendTaskContext,
		updateTaskContext,
		customFields,
		service.eventRunTask,
		service.eventUpdateDependentKeyWord,
	)
}

func (service *CollectorService) createTriggerRepositoryByName(taskAppService itask.ITask, repository dataModel.RepositoryModel) (task itask.ITask, err error) {
	var (
		taskKey           string
		updateTaskContext = dataModel.RepositoryModel{}
		customFields      = taskAppService
		sendTaskContext   *contextTaskSend
		uniqueKey         = repository.Name
	)
	if taskKey, err = service.createKeyForTask(
		CompositeByName,
		taskAppService,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	taskKey = strings.Join([]string{"(trigger)", taskKey}, " ")
	if sendTaskContext, err = service.createSendContextForTask(
		OnlyDescriptions,
		taskKey,
		[]dataModel.RepositoryModel{repository},
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		CompositeByName,
		taskKey,
		sendTaskContext,
		updateTaskContext,
		customFields,
		service.eventRunTask,
		service.eventUpdateTriggerByName,
	)
}

func (service *CollectorService) createDependentIssuesByName(triggerTask itask.ITask, repository dataModel.RepositoryModel) (constructor itask.ITask, err error) {
	var (
		taskKey           string
		customFields      = triggerTask
		sendTaskContext   *contextTaskSend
		updateTaskContext = make([]dataModel.IssueModel, 0)
		uniqueKey         = repository.Name
	)
	if taskKey, err = service.createKeyForTask(
		CompositeByName,
		triggerTask.GetState().GetCustomFields().(itask.ITask),
		uniqueKey,
	); err != nil {
		return nil, err
	}
	taskKey = strings.Join([]string{"(dependent)", taskKey}, " ")
	if sendTaskContext, err = service.createSendContextForTask(
		OnlyIssues,
		taskKey,
		repository,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		CompositeByName,
		taskKey,
		sendTaskContext,
		updateTaskContext,
		customFields,
		service.eventRunTask,
		service.eventUpdateDependentByName,
	)
}
