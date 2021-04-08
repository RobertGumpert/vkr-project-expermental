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
		taskAppService := task.GetState().GetCustomFields().(itask.ITask)
		repositories := taskAppService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		runtimeinfo.LogInfo("TASK COMPLETED [", task.GetKey(), "] LEN. [", len(repositories), "]")
		deleteTasks[task.GetKey()] = struct{}{}
		taskAppService.GetState().GetCustomFields().(chan itask.ITask) <- taskAppService
		break
	case RepositoryIssues:
		taskAppService := task.GetState().GetCustomFields().(itask.ITask)
		repositoryID := taskAppService.GetState().GetCustomFields().(uint)
		issues := taskAppService.GetState().GetUpdateContext().([]dataModel.IssueModel)
		runtimeinfo.LogInfo("TASK COMPLETED [", task.GetKey(), "] LEN. [", len(issues), "] FOR REPO [", repositoryID, "]")
		deleteTasks[task.GetKey()] = struct{}{}
		taskAppService.GetState().GetCustomFields().(chan itask.ITask) <- taskAppService
		break
	case RepositoriesDescriptionAndIssues:
		var (
			taskAppService      itask.ITask
			isDependent, trigger = task.IsDependent()
		)
		if isDependent {
			isCompleted, dependentTasks, err := service.taskManager.TriggerIsCompleted(trigger)
			if err != nil {
				runtimeinfo.LogError("TASK COMPLETED WITH ERROR: {", err, "} [", task.GetKey(), "]")
				return nil
			}
			if isCompleted {
				taskAppService = trigger.GetState().GetCustomFields().(itask.ITask)
				repository := trigger.GetState().GetUpdateContext().(dataModel.RepositoryModel)
				issues := task.GetState().GetUpdateContext().([]dataModel.IssueModel)
				repository.Issues = append(repository.Issues, issues...)
				taskGateRepositories := taskAppService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
				taskGateRepositories = append(taskGateRepositories, repository)
				taskAppService.GetState().SetUpdateContext(taskGateRepositories)
				for dependentTaskKey := range dependentTasks {
					deleteTasks[dependentTaskKey] = struct{}{}
				}
				deleteTasks[trigger.GetKey()] = struct{}{}

			}
			if taskAppService != nil {
				sendContext := taskAppService.GetState().GetSendContext().([]dataModel.RepositoryModel)
				updateContext := taskAppService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
				if len(sendContext) == len(updateContext) {
					taskAppService.GetState().GetCustomFields().(chan itask.ITask) <- taskAppService
				}
			}
		}
		break
	}
	return deleteTasks
}


func (service *CollectorService) writeToGateTaskDownloadedRepositories() {

}