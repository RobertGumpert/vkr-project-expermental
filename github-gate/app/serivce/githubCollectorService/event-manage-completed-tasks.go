package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
)

func (service *CollectorService) eventManageCompletedTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	deleteTasks = make(map[string]struct{})
	switch task.GetType() {
	case RepositoriesDescription:
		taskGateService := task.GetState().GetCustomFields().(itask.ITask)
		repositories := taskGateService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		runtimeinfo.LogInfo("TASK COMPLETED [", task.GetKey(), "] LEN. [", len(repositories), "]")
		deleteTasks[task.GetKey()] = struct{}{}
		taskGateService.GetState().GetCustomFields().(chan itask.ITask) <- taskGateService
		break
	case RepositoryIssues:
		taskGateService := task.GetState().GetCustomFields().(itask.ITask)
		repositoryID := taskGateService.GetState().GetCustomFields().(uint)
		issues := taskGateService.GetState().GetUpdateContext().([]dataModel.IssueModel)
		runtimeinfo.LogInfo("TASK COMPLETED [", task.GetKey(), "] LEN. [", len(issues), "] FOR REPO [", repositoryID, "]")
		deleteTasks[task.GetKey()] = struct{}{}
		taskGateService.GetState().GetCustomFields().(chan itask.ITask) <- taskGateService
		break
	case RepositoriesDescriptionAndIssues:
		var (
			taskGateService      itask.ITask
			isDependent, trigger = task.IsDependent()
		)
		if isDependent {
			isCompleted, dependentTasks, err := service.taskManager.TriggerIsCompleted(trigger)
			if err != nil {
				runtimeinfo.LogError("TASK COMPLETED WITH ERROR: {", err, "} [", task.GetKey(), "]")
				return nil
			}
			if isCompleted {
				taskGateService = trigger.GetState().GetCustomFields().(itask.ITask)
				repository := trigger.GetState().GetUpdateContext().(dataModel.RepositoryModel)
				issues := task.GetState().GetUpdateContext().([]dataModel.IssueModel)
				repository.Issues = append(repository.Issues, issues...)
				taskGateRepositories := taskGateService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
				taskGateRepositories = append(taskGateRepositories, repository)
				taskGateService.GetState().SetUpdateContext(taskGateRepositories)
				for dependentTaskKey := range dependentTasks {
					deleteTasks[dependentTaskKey] = struct{}{}
				}
				deleteTasks[trigger.GetKey()] = struct{}{}

			}
			if taskGateService != nil {
				sendContext := taskGateService.GetState().GetSendContext().([]dataModel.RepositoryModel)
				updateContext := taskGateService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
				if len(sendContext) == len(updateContext) {
					taskGateService.GetState().GetCustomFields().(chan itask.ITask) <- taskGateService
				}
			}
		}
		if taskGateService == nil {
			runtimeinfo.LogError("TASK COMPLETED WITH ERROR: {NONE EXIST TASK FROM APP SERVICE} [", task.GetKey(), "]")
		}
		break
	}
	return deleteTasks
}
