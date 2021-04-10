package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
)

func (service *CollectorService) eventManageCompletedTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	deleteTasks = make(map[string]struct{})
	switch task.GetType() {
	case OnlyDescriptions:
		taskAppService := task.GetState().GetCustomFields().(itask.ITask)
		repositories := taskAppService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		runtimeinfo.LogInfo("TASK COMPLETED [", task.GetKey(), "] LEN. [", len(repositories), "]")
		deleteTasks[task.GetKey()] = struct{}{}
		taskAppService.GetState().GetCustomFields().(chan itask.ITask) <- taskAppService
		break
	case OnlyIssues:
		taskAppService := task.GetState().GetCustomFields().(itask.ITask)
		repositoryID := taskAppService.GetState().GetCustomFields().(uint)
		issues := taskAppService.GetState().GetUpdateContext().([]dataModel.IssueModel)
		runtimeinfo.LogInfo("TASK COMPLETED [", task.GetKey(), "] LEN. [", len(issues), "] FOR REPO [", repositoryID, "]")
		deleteTasks[task.GetKey()] = struct{}{}
		taskAppService.GetState().GetCustomFields().(chan itask.ITask) <- taskAppService
		break
	case CompositeByKeyWord:
		var (
			triggerIsCompleted bool
			trigger            itask.ITask
		)
		deleteTasks, trigger, triggerIsCompleted = service.manageCompletedTaskRepositoriesByKeyWord(task)
		if triggerIsCompleted {
			deleteTasks[trigger.GetKey()] = struct{}{}
			repositories := trigger.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
			taskAppService := trigger.GetState().GetCustomFields().(itask.ITask)
			taskAppService.GetState().SetUpdateContext(repositories)
			taskAppService.GetState().GetCustomFields().(chan itask.ITask) <- taskAppService
		} else {
			deleteTasks = nil
		}
		break
	case CompositeByName:
		deleteTasks = service.manageCompletedTaskRepositoryByName(task)
		break
	case RepositoryAndRepositoriesContainingKeyWord:
		deleteTasks = service.manageCompletedRepositoryAndRepositoriesByKeyWord(task)
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

func (service *CollectorService) manageCompletedTaskRepositoriesByKeyWord(task itask.ITask) (deleteTasks map[string]struct{}, trigger itask.ITask, triggerIsCompleted bool) {
	deleteTasks = make(map[string]struct{})
	var (
		countCompletedDependentTasks int
		isDependent, isTrigger       = false, false
	)
	triggerIsCompleted = false
	isDependent, trigger = task.IsDependent()
	isTrigger, _ = task.IsTrigger()
	if isDependent && !isTrigger {
		_, dependentsTasks := trigger.IsTrigger()
		repositories := trigger.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		for _, dependentTask := range *dependentsTasks {
			if dependentTask.GetState().IsCompleted() {
				countCompletedDependentTasks++
			}
			deleteTasks[dependentTask.GetKey()] = struct{}{}
		}
		if len(repositories) == countCompletedDependentTasks {
			triggerIsCompleted = true
		}
	}
	return deleteTasks, trigger, triggerIsCompleted
}

func (service *CollectorService) manageCompletedRepositoryAndRepositoriesByKeyWord(task itask.ITask) (deleteTasks map[string]struct{}) {
	deleteTasks = make(map[string]struct{})
	customFields := task.GetState().GetCustomFields().(*compositeCustomFields)
	switch customFields.TaskType {
	case CompositeByKeyWord:
		var (
			countCompletedDependentTasks     int
			triggerIsCompleted               bool
			triggerSearchByKeyWord           itask.ITask
			repositoryTrigger                itask.ITask
			repositoryTriggerDependentsTasks *[]itask.ITask
		)
		deleteTasks, triggerSearchByKeyWord, triggerIsCompleted = service.manageCompletedTaskRepositoriesByKeyWord(task)
		if triggerIsCompleted {
			updateContext := triggerSearchByKeyWord.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
			taskAppService := triggerSearchByKeyWord.GetState().GetCustomFields().(itask.ITask)
			repositories := taskAppService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
			repositories = append(repositories, updateContext...)
			taskAppService.GetState().SetUpdateContext(repositories)
			_, repositoryTrigger = triggerSearchByKeyWord.IsDependent()
			_, repositoryTriggerDependentsTasks = repositoryTrigger.IsTrigger()
			for _, dependentTask := range *repositoryTriggerDependentsTasks {
				if dependentTask.GetState().IsCompleted() {
					countCompletedDependentTasks++
				}
				deleteTasks[dependentTask.GetKey()] = struct{}{}
			}
			if countCompletedDependentTasks == len(*repositoryTriggerDependentsTasks) {
				deleteTasks[repositoryTrigger.GetKey()] = struct{}{}
				taskAppService.GetState().GetCustomFields().(chan itask.ITask) <- taskAppService
			} else {
				deleteTasks = nil
			}
		} else {
			deleteTasks = nil
		}
		break
	case OnlyDescriptions:
		updateContext := task.GetState().GetUpdateContext().(dataModel.RepositoryModel)
		taskAppService := customFields.Fields.(itask.ITask)
		repositories := taskAppService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		repositories = append(repositories, updateContext)
		taskAppService.GetState().SetUpdateContext(repositories)
		deleteTasks = nil
		break
	}
	return deleteTasks
}
