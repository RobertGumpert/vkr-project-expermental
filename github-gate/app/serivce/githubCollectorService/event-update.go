package githubCollectorService

import (
	"errors"
	"fmt"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
)

func (service *CollectorService) eventUpdateOnlyDescriptions(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
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

func (service *CollectorService) eventUpdateOnlyIssues(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
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

func (service *CollectorService) eventUpdateTriggerByName(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	cast := somethingUpdateContext.(*jsonSendFromCollectorDescriptionsRepositories)
	models := service.writeRepositoriesToDB(cast.Repositories)
	if cast.ExecutionTaskStatus.TaskCompleted {
		task.GetState().SetCompleted(true)
		task.GetState().SetUpdateContext(models[0])
	}
	return nil, false
}

func (service *CollectorService) eventUpdateDependentByName(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
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

func (service *CollectorService) eventUpdateTriggerByKeyWord(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	cast := somethingUpdateContext.(*jsonSendFromCollectorRepositoriesByKeyWord)
	models := service.writeRepositoriesToDB(cast.Repositories)
	isTrigger, dependentsTasks := task.IsTrigger()
	if cast.ExecutionTaskStatus.TaskCompleted {
		if isTrigger {
			var (
				deleteTasksKeys      = make(map[string]struct{})
				deleteDependentTasks []itask.ITask
				next                 = 0
			)
			for ; next < len(models); next++ {
				model := models[next]
				dependent := (*dependentsTasks)[next]
				customFields := dependent.GetState().GetCustomFields().(*compositeCustomFields)
				sendContext := dependent.GetState().GetSendContext().(*contextTaskSend)
				updateKey := fmt.Sprintf(dependent.GetKey(), model.Name)
				sendContext.JSONBody.(*jsonSendToCollectorRepositoryIssues).TaskKey = updateKey
				sendContext.JSONBody.(*jsonSendToCollectorRepositoryIssues).Repository.Name = model.Name
				sendContext.JSONBody.(*jsonSendToCollectorRepositoryIssues).Repository.Owner = model.Owner
				dependent.GetState().SetSendContext(sendContext)
				customFields.Fields = model
				dependent.GetState().SetCustomFields(customFields)
				dependent.SetKey(updateKey)
			}
			deleteDependentTasks = (*dependentsTasks)[next:]
			for _, dependent := range deleteDependentTasks {
				deleteTasksKeys[dependent.GetKey()] = struct{}{}
			}
			service.taskManager.DeleteTasksByKeys(deleteTasksKeys)
			*dependentsTasks = (*dependentsTasks)[:next]
		} else {
			return errors.New("Isn't trigger. "), true
		}
		task.GetState().SetUpdateContext(models)
		task.GetState().SetCompleted(true)
	}
	return nil, false
}

func (service *CollectorService) eventUpdateDependentKeyWord(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	cast := somethingUpdateContext.(*jsonSendFromCollectorRepositoryIssues)
	repository := task.GetState().GetCustomFields().(*compositeCustomFields).Fields.(dataModel.RepositoryModel)
	if repository.ID == 0 {
		err = errors.New("Repository ID is 0. ")
		runtimeinfo.LogError(err)
		return err, true
	}
	_ = service.writeIssuesToDB(cast.Issues, repository.ID)
	if cast.ExecutionTaskStatus.TaskCompleted {
		task.GetState().SetCompleted(true)
	}
	return nil, false
}

func (service *CollectorService) eventUpdateTriggerRepositoryAndRepositoriesKeyWord(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	cast := somethingUpdateContext.(*jsonSendFromCollectorDescriptionsRepositories)
	if cast.ExecutionTaskStatus.TaskCompleted {
		var (
			updateContext dataModel.RepositoryModel
		)
		if isTrigger, dependentsTasks := task.IsTrigger(); isTrigger {
			updateContext = service.writeRepositoriesToDB(cast.Repositories)[0]
			for _, dependentTask := range *dependentsTasks {
				customFields := dependentTask.GetState().GetCustomFields().(*compositeCustomFields)
				if customFields.TaskType == TaskTypeDownloadOnlyIssues {
					customFields.Fields = updateContext
					dependentTask.GetState().SetCustomFields(customFields)
					service.taskManager.TakeOffRunBanInQueue(dependentTask)
				}
			}
		} else {
			return errors.New("Isn't trigger. "), true
		}
		task.GetState().SetCompleted(true)
		task.GetState().SetUpdateContext(updateContext)
	}
	return nil, false
}

func (service *CollectorService) eventUpdateDependentRepositoryAndRepositoriesKeyWord(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	customFields := task.GetState().GetCustomFields().(*compositeCustomFields)
	runtimeinfo.LogInfo(customFields.TaskType)
	switch customFields.TaskType {
	case TaskTypeDownloadCompositeByKeyWord:
		var (
			isTrigger, _   = task.IsTrigger()
			isDependent, _ = task.IsDependent()
		)
		if isTrigger && isDependent {
			return service.eventUpdateTriggerByKeyWord(task, somethingUpdateContext)
		}
		if !isTrigger && isDependent {
			return service.eventUpdateDependentKeyWord(task, somethingUpdateContext)
		}
		break
	case TaskTypeDownloadOnlyIssues:
		cast := somethingUpdateContext.(*jsonSendFromCollectorRepositoryIssues)
		customFields := task.GetState().GetCustomFields().(*compositeCustomFields)
		repositoryID := customFields.Fields.(dataModel.RepositoryModel).ID
		models := service.writeIssuesToDB(cast.Issues, repositoryID)
		if cast.ExecutionTaskStatus.TaskCompleted {
			task.GetState().SetCompleted(true)
			_, trigger := task.IsDependent()
			_, dependentsTasks := trigger.IsTrigger()
			for _, dependentTask := range *dependentsTasks {
				customFields := dependentTask.GetState().GetCustomFields().(*compositeCustomFields)
				if customFields.TaskType == TaskTypeDownloadCompositeByKeyWord {
					service.taskManager.TakeOffRunBanInQueue(dependentTask)
				}
			}
		}
		updateContext := task.GetState().GetUpdateContext().([]dataModel.IssueModel)
		updateContext = append(updateContext, models...)
		task.GetState().SetUpdateContext(updateContext)
		break
	}
	return nil, false
}
