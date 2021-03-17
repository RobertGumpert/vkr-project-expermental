package githubTasksService

import (
	"errors"
	"github-gate/app/models/interapplicationModels/githubCollectorModels"
	"github-gate/pckg/runtimeinfo"
	"strings"
)

func (service *GithubTasksService) updateTaskRepositoriesDescriptionByURL(updateStateTask *githubCollectorModels.UpdateTaskRepositoriesByURLS) error {
	var (
		taskForCollector *TaskForCollector
		taskKey          = updateStateTask.ExecutionTaskStatus.TaskKey
		executionStatus  = updateStateTask.ExecutionTaskStatus.TaskCompleted
		listRepositories = updateStateTask.Repositories
	)
	if strings.TrimSpace(taskKey) == "" {
		return errors.New("GETTING FROM GITHUB-COLLECTOR TASK WITH EMPTY KEY. ")
	}
	for i := 0; i < len(service.tasksForCollectorsQueue); i++ {
		if service.tasksForCollectorsQueue[i].GetKey() == taskKey {
			taskForCollector = service.tasksForCollectorsQueue[i]
			break
		}
	}
	if taskForCollector == nil {
		return errors.New("GETTING FROM GITHUB-COLLECTOR TASK WITH KEY [" + taskKey + "] ISN'T EXIST. ")
	}
	runtimeinfo.LogInfo("GETTING UPDATE (status: ", executionStatus, ") FOR TASK [", taskKey, "] WITH LIST ELEMENTS SIZE OF [", len(listRepositories), "]")
	//for i := 0; i < len(updateStateTask.Repositories); i++ {
	//
	//}
	if executionStatus {
		service.completedTasksChannel <- taskForCollector
	}
	return nil
}
