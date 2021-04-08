package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
)

func (service *CollectorService) eventManageCompletedTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	deleteTasks = make(map[string]struct{})
	switch task.GetType() {
	case RepositoriesOnlyDescription:
		taskAppService := task.GetState().GetCustomFields().(itask.ITask)
		repositories := taskAppService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		runtimeinfo.LogInfo("TASK COMPLETED [", task.GetKey(), "] LEN. [", len(repositories), "]")
		deleteTasks[task.GetKey()] = struct{}{}
		taskAppService.GetState().GetCustomFields().(chan itask.ITask) <- taskAppService
		break
	case RepositoryOnlyIssues:
		taskAppService := task.GetState().GetCustomFields().(itask.ITask)
		repositoryID := taskAppService.GetState().GetCustomFields().(uint)
		issues := taskAppService.GetState().GetUpdateContext().([]dataModel.IssueModel)
		runtimeinfo.LogInfo("TASK COMPLETED [", task.GetKey(), "] LEN. [", len(issues), "] FOR REPO [", repositoryID, "]")
		deleteTasks[task.GetKey()] = struct{}{}
		taskAppService.GetState().GetCustomFields().(chan itask.ITask) <- taskAppService
		break
	case RepositoriesByKeyWord:
		break
	case RepositoryByName:
		deleteTasks = service.manageCompletedTaskRepositoryByName(task)
		break
	}
	return deleteTasks
}

func (service *CollectorService) manageCompletedTaskRepositoryByName(task itask.ITask) (deleteTasks map[string]struct{}) {
	deleteTasks = make(map[string]struct{})
	var (
		taskAppService       itask.ITask
		isDependent, trigger = task.IsDependent()
	)
	if isDependent {
		triggerIsCompleted, dependentTasks, err := service.taskManager.TriggerIsCompleted(trigger)
		if err != nil {
			runtimeinfo.LogError("TASK COMPLETED WITH ERROR: {", err, "} [", task.GetKey(), "]")
			return nil
		}
		if triggerIsCompleted {
			taskAppService = trigger.GetState().GetCustomFields().(itask.ITask)
			triggerUpdateContext := trigger.GetState().GetUpdateContext().(dataModel.RepositoryModel)
			dependentUpdateContext := task.GetState().GetUpdateContext().([]dataModel.IssueModel)
			triggerUpdateContext.Issues = append(triggerUpdateContext.Issues, dependentUpdateContext...)
			taskAppUpdateContext := taskAppService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
			taskAppUpdateContext = append(taskAppUpdateContext, triggerUpdateContext)
			taskAppService.GetState().SetUpdateContext(taskAppUpdateContext)
			for dependentTaskKey := range dependentTasks {
				deleteTasks[dependentTaskKey] = struct{}{}
			}
			deleteTasks[trigger.GetKey()] = struct{}{}
			taskAppService.GetState().GetCustomFields().(chan itask.ITask) <- taskAppService
		}
	}
	return deleteTasks
}

func (service *CollectorService) manageCompletedTaskRepositoriesByKeyWord(task itask.ITask) (deleteTasks map[string]struct{}) {
	deleteTasks = make(map[string]struct{})
	var (
		countCompletedDependentTasks int
		triggerIsCompleted           = false
		taskAppService               itask.ITask
		isDependent, trigger         = task.IsDependent()
	)
	if isDependent {
		_, dependentsTasks := trigger.IsTrigger()
		repositories := trigger.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		for _, dependentTask := range dependentsTasks {
			if dependentTask.GetState().IsCompleted() {
				countCompletedDependentTasks++
			}
			deleteTasks[dependentTask.GetKey()] = struct{}{}
		}
		if len(repositories) == countCompletedDependentTasks {
			deleteTasks[trigger.GetKey()] = struct{}{}
			triggerIsCompleted = true
		}
		if triggerIsCompleted {
			repositories := trigger.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
			taskAppService = trigger.GetState().GetCustomFields().(itask.ITask)
			taskAppService.GetState().SetUpdateContext(repositories)
			taskAppService.GetState().GetCustomFields().(chan itask.ITask) <- taskAppService
		} else {
			deleteTasks = nil
		}
	}
	return deleteTasks
}