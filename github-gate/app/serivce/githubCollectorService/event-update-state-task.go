package githubCollectorService

import (
	"errors"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
)

func (service *CollectorService) eventUpdateTaskDescriptionsRepositories(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	cast := somethingUpdateContext.(*jsonSendFromCollectorDescriptionsRepositories)
	models := service.writeRepositoriesToDB(cast.Repositories)
	if cast.ExecutionTaskStatus.TaskCompleted {
		task.GetState().SetCompleted(true)
	}
	gateServiceTask := task.GetState().GetCustomFields().(itask.ITask)
	updateContext := gateServiceTask.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
	updateContext = append(updateContext, models...)
	gateServiceTask.GetState().SetUpdateContext(updateContext)
	return nil, false
}

func (service *CollectorService) eventUpdateTaskRepositoryIssues(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	cast := somethingUpdateContext.(*jsonSendFromCollectorRepositoryIssues)
	gateServiceTask := task.GetState().GetCustomFields().(itask.ITask)
	repositoryID := gateServiceTask.GetState().GetCustomFields().(uint)
	models := service.writeIssuesToDB(cast.Issues, repositoryID)
	if cast.ExecutionTaskStatus.TaskCompleted {
		task.GetState().SetCompleted(true)
	}
	updateContext := gateServiceTask.GetState().GetUpdateContext().([]dataModel.IssueModel)
	updateContext = append(updateContext, models...)
	gateServiceTask.GetState().SetUpdateContext(updateContext)
	return nil, false
}

func (service *CollectorService) eventUpdateTriggerDescriptionRepository(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	cast := somethingUpdateContext.(*jsonSendFromCollectorDescriptionsRepositories)
	models := service.writeRepositoriesToDB(cast.Repositories)
	if cast.ExecutionTaskStatus.TaskCompleted {
		task.GetState().SetCompleted(true)
		task.GetState().SetUpdateContext(models[0])
	}
	return nil, false
}

func (service *CollectorService) eventUpdateDependentRepositoryIssues(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	cast := somethingUpdateContext.(*jsonSendFromCollectorRepositoryIssues)
	triggerTask := task.GetState().GetCustomFields().(itask.ITask)
	repository := triggerTask.GetState().GetUpdateContext().(dataModel.RepositoryModel)
	if repository.ID == 0 {
		return errors.New("Repository ID is 0. "), true
	}
	models := service.writeIssuesToDB(cast.Issues, repository.ID)
	if cast.ExecutionTaskStatus.TaskCompleted {
		task.GetState().SetCompleted(true)
	}
	updateContext := task.GetState().GetUpdateContext().([]dataModel.IssueModel)
	updateContext = append(updateContext, models...)
	task.GetState().SetUpdateContext(updateContext)
	return nil, false
}
