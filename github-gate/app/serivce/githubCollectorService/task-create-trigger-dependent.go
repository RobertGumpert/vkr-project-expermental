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
		RepositoriesByKeyWord,
		taskAppService,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	taskKey = strings.Join([]string{"(trigger)", taskKey}, " ")
	if sendTaskContext, err = service.createSendContextForTask(
		RepositoriesByKeyWord,
		taskKey,
		keyWord,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		RepositoriesByKeyWord,
		taskKey,
		sendTaskContext,
		updateTaskContext,
		customFields,
		service.eventRunTask,
		service.eventUpdateTriggerRepositoriesByKeyWord,
	)
}

func (service *CollectorService) createDependentIssuesByKeyWord(triggerTask itask.ITask, number int, keyWord string) (constructor itask.ITask, err error) {
	var (
		taskKey           string
		sendTaskContext   *contextTaskSend
		customFields      = dataModel.RepositoryModel{}
		updateTaskContext = make([]dataModel.IssueModel, 0)
		uniqueKey         = "%s"
	)
	if taskKey, err = service.createKeyForTask(
		RepositoriesByKeyWord,
		triggerTask.GetState().GetCustomFields().(itask.ITask),
		uniqueKey,
	); err != nil {
		return nil, err
	}
	dependentNumberKey := strings.Join([]string{"(dependent-", strconv.Itoa(number), ")"}, "")
	taskKey = strings.Join([]string{dependentNumberKey, taskKey}, " ")
	if sendTaskContext, err = service.createSendContextForTask(
		RepositoryOnlyIssues,
		taskKey,
		dataModel.RepositoryModel{
			Name:  keyWord,
			Owner: keyWord,
		},
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		RepositoriesByKeyWord,
		taskKey,
		sendTaskContext,
		updateTaskContext,
		customFields,
		service.eventRunTask,
		service.eventUpdateDependentIssuesByKeyWord,
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
		RepositoryByName,
		taskAppService,
		uniqueKey,
	); err != nil {
		return nil, err
	}
	taskKey = strings.Join([]string{"(trigger)", taskKey}, " ")
	if sendTaskContext, err = service.createSendContextForTask(
		RepositoriesOnlyDescription,
		taskKey,
		[]dataModel.RepositoryModel{repository},
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		RepositoryByName,
		taskKey,
		sendTaskContext,
		updateTaskContext,
		customFields,
		service.eventRunTask,
		service.eventUpdateTriggerRepositoryByName,
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
		RepositoryByName,
		triggerTask.GetState().GetCustomFields().(itask.ITask),
		uniqueKey,
	); err != nil {
		return nil, err
	}
	taskKey = strings.Join([]string{"(dependent)", taskKey}, " ")
	if sendTaskContext, err = service.createSendContextForTask(
		RepositoryOnlyIssues,
		taskKey,
		repository,
	); err != nil {
		return nil, err
	}
	return service.taskManager.CreateTask(
		RepositoryByName,
		taskKey,
		sendTaskContext,
		updateTaskContext,
		customFields,
		service.eventRunTask,
		service.eventUpdateDependentIssuesByName,
	)
}
