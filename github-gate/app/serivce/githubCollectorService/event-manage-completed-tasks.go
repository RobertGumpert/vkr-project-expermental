package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"strconv"
)

func (service *CollectorService) eventManageCompletedTasks(task itask.ITask) (deleteTasks, saveTasks map[string]struct{}) {
	deleteTasks, saveTasks = make(map[string]struct{}), make(map[string]struct{})
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
	case RepositoriesDescriptionAndIssues:
		var (
			taskGateService      itask.ITask
			isDependent, trigger = task.IsDependent()
		)
		if isDependent {
			isCompleted, dependentTasksCompletedFlags, err := service.taskSteward.TriggerIsCompleted(trigger)
			if err != nil {
				runtimeinfo.LogError("TASK COMPLETED [", task.GetKey(), "] WITH ERROR: ", err)
				return nil, nil
			}
			if isCompleted {
				taskGateService = trigger.GetState().GetCustomFields().(itask.ITask)
				repository := trigger.GetState().GetUpdateContext().(dataModel.RepositoryModel)
				issues := task.GetState().GetUpdateContext().([]dataModel.IssueModel)
				repository.Issues = append(repository.Issues, issues...)
				taskGateRepositories := taskGateService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
				taskGateRepositories = append(taskGateRepositories, repository)
				taskGateService.GetState().SetUpdateContext(taskGateRepositories)
				for dependentTaskKey := range dependentTasksCompletedFlags {
					deleteTasks[dependentTaskKey] = struct{}{}
				}
				deleteTasks[trigger.GetKey()] = struct{}{}
			}
		}
		if taskGateService != nil {
			updateContext := taskGateService.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
			if len(taskGateService.GetState().GetSendContext().([]dataModel.RepositoryModel)) ==
				len(updateContext) {
				taskGateService.GetState().SetCompleted(true)
				for _, update := range updateContext {
					runtimeinfo.LogInfo("TASK TRIGGER COMPLETED [", trigger.GetKey(), "] FOR: ", update.Name, " WITH ISSUES LIST SIZE OF ", strconv.Itoa(len(update.Issues)))
				}
			}
		}
		break
	}
	return deleteTasks, saveTasks
}
