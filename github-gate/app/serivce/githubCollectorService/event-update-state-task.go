package githubCollectorService

import (
	"errors"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
)

func (service *CollectorService) eventUpdateTaskDescriptionsRepositories(task itask.ITask, interProgramUpdateContext interface{}) (err error) {
	cast := interProgramUpdateContext.(*jsonSendFromCollectorDescriptionsRepositories)
	models := service.writeRepositoriesToDB(cast.Repositories)
	if cast.ExecutionTaskStatus.TaskCompleted {
		task.GetState().SetCompleted(true)
	}
	gateServiceTask := task.GetState().GetCustomFields().(itask.ITask)
	updateContext := gateServiceTask.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
	updateContext = append(updateContext, models...)
	gateServiceTask.GetState().SetUpdateContext(updateContext)
	return nil
}

func (service *CollectorService) eventUpdateTaskRepositoryIssues(task itask.ITask, interProgramUpdateContext interface{}) (err error) {
	cast := interProgramUpdateContext.(*jsonSendFromCollectorRepositoryIssues)
	gateServiceTask := task.GetState().GetCustomFields().(itask.ITask)
	repositoryID := gateServiceTask.GetState().GetCustomFields().(uint)
	models := service.writeIssuesToDB(cast.Issues, repositoryID)
	if cast.ExecutionTaskStatus.TaskCompleted {
		task.GetState().SetCompleted(true)
	}
	updateContext := gateServiceTask.GetState().GetUpdateContext().([]dataModel.IssueModel)
	updateContext = append(updateContext, models...)
	gateServiceTask.GetState().SetUpdateContext(updateContext)
	return nil
}

func (service *CollectorService) eventUpdateTriggerDescriptionRepository(task itask.ITask, interProgramUpdateContext interface{}) (err error) {
	cast := interProgramUpdateContext.(*jsonSendFromCollectorDescriptionsRepositories)
	models := service.writeRepositoriesToDB(cast.Repositories)
	if cast.ExecutionTaskStatus.TaskCompleted {
		task.GetState().SetCompleted(true)
		task.GetState().SetUpdateContext(models[0])
	}
	return nil
}

func (service *CollectorService) eventUpdateDependentRepositoryIssues(task itask.ITask, interProgramUpdateContext interface{}) (err error) {
	cast := interProgramUpdateContext.(*jsonSendFromCollectorRepositoryIssues)
	triggerTask := task.GetState().GetCustomFields().(itask.ITask)
	if !triggerTask.GetState().IsCompleted() {
		return errors.New("Trigger [" + triggerTask.GetKey() + "] isn't completed. ")
	}
	repository := triggerTask.GetState().GetUpdateContext().(dataModel.RepositoryModel)
	if repository.ID == 0 {
		return errors.New("Repository ID is 0. ")
	}
	models := service.writeIssuesToDB(cast.Issues, repository.ID)
	if cast.ExecutionTaskStatus.TaskCompleted {
		task.GetState().SetCompleted(true)
	}
	updateContext := task.GetState().GetUpdateContext().([]dataModel.IssueModel)
	updateContext = append(updateContext, models...)
	task.GetState().SetUpdateContext(updateContext)
	return nil
}
