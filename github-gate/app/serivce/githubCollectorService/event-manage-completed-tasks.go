package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
)

func (service *CollectorService) eventManageCompletedTasks(task itask.ITask) (deleteTasks, saveTasks map[string]struct{}) {
	deleteTasks, saveTasks = make(map[string]struct{}), make(map[string]struct{})
	var (
		isTrigger, dependents = task.IsTrigger()
		isDependent, trigger  = task.IsDependent()
	)
	if !isTrigger && !isDependent {
		switch task.GetType() {
		case RepositoriesDescription:
			taskGateService := task.GetState().GetCustomFields().(itask.ITask)
			repositories := taskGateService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
			runtimeinfo.LogInfo("TASK COMPLETED [", task.GetKey(), "] LEN. [", len(repositories), "]")
			deleteTasks[task.GetKey()] = struct{}{}
			break
		case RepositoryIssues:
			taskGateService := task.GetState().GetCustomFields().(itask.ITask)
			repositoryID := taskGateService.GetState().GetCustomFields().(uint)
			issues := taskGateService.GetState().GetUpdateContext().([]dataModel.IssueModel)
			runtimeinfo.LogInfo("TASK COMPLETED [", task.GetKey(), "] LEN. [", len(issues), "] FOR REPO [", repositoryID, "]")
			deleteTasks[task.GetKey()] = struct{}{}
			break
		}
	} else {
		if isDependent {
			countCompletedDecedentTasks := 0
			_, dependents = trigger.IsTrigger()
			for _, dependent := range dependents {
				if dependent.GetState().IsCompleted() {
					deleteTasks[dependent.GetKey()] = struct{}{}
					countCompletedDecedentTasks++
				}
			}
			if countCompletedDecedentTasks == len(dependents) {
				deleteTasks[trigger.GetKey()] = struct{}{}
			} else {
				deleteTasks = nil
			}
		}
	}
	return deleteTasks, saveTasks
}
