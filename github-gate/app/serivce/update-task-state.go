package serivce

import (
	"errors"
	"github-gate/app/models/updateTaskModel"
	"github-gate/pckg/runtimeinfo"
)

func (a *AppService) UpdateStateTaskRepositoriesByURL(updateTaskState *updateTaskModel.RepositoriesByURLS) error {
	key := updateTaskState.ExecutionTaskStatus.TaskKey
	if value, exist := a.tasks.Get(key); !exist {
		err := errors.New("task with key [" + key + "] isn't exist ")
		runtimeinfo.LogError(err)
		return err
	} else {
		task := value.(*task)
		task.ExecutionStatus = updateTaskState.ExecutionTaskStatus.TaskCompleted
		if task.ExecutionStatus {
			a.tasksChannel <- task.TaskKey
		}
		runtimeinfo.LogInfo("UPDATE TASK with key :[", task.TaskKey, "] is competed ;[", task.ExecutionStatus, "] count elements [", len(updateTaskState.Repositories), "]")
		//for _, repo := range updateTaskState.Repositories {
		//	repositoryViewModel := &ViewModelRepository{
		//		URL:         repo.URL,
		//		Topics:      repo.Topics,
		//		Description: repo.Description,
		//	}
		//	str := fmt.Sprintf(
		//		"URL : %s\n\t\tTOPICS : [%s]\n\t\tABOUT : [%s]\n\t\tERR : [%s]",
		//		repositoryViewModel.URL,
		//		strings.Join(repositoryViewModel.Topics, ", "),
		//		repositoryViewModel.Description,
		//		repo.Err,
		//	)
		//	fmt.Println(str)
		//}
	}
	return nil
}

func (a *AppService) UpdateStateTaskRepositoryIssues(updateTaskState *updateTaskModel.RepositoryIssues) error {
	key := updateTaskState.ExecutionTaskStatus.TaskKey
	if value, exist := a.tasks.Get(key); !exist {
		err := errors.New("task with key [" + key + "] isn't exist ")
		runtimeinfo.LogError(err)
		return err
	} else {
		task := value.(*task)
		task.ExecutionStatus = updateTaskState.ExecutionTaskStatus.TaskCompleted
		results := task.Results.(int)
		results = results + len(updateTaskState.Issues)
		task.Results = results
		if task.ExecutionStatus {
			a.tasksChannel <- task.TaskKey
		}
		runtimeinfo.LogInfo("UPDATE TASK with key :[", task.TaskKey, "] is competed ;[", task.ExecutionStatus, "] count elements [", len(updateTaskState.Issues), "]")
		//for _, repo := range updateTaskState.Repositories {
		//	repositoryViewModel := &ViewModelRepository{
		//		URL:         repo.URL,
		//		Topics:      repo.Topics,
		//		Description: repo.Description,
		//	}
		//	str := fmt.Sprintf(
		//		"URL : %s\n\t\tTOPICS : [%s]\n\t\tABOUT : [%s]\n\t\tERR : [%s]",
		//		repositoryViewModel.URL,
		//		strings.Join(repositoryViewModel.Topics, ", "),
		//		repositoryViewModel.Description,
		//		repo.Err,
		//	)
		//	fmt.Println(str)
		//}
	}
	return nil
}

