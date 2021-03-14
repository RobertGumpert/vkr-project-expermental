package tasksService

import (
	"issue-indexer/app/models/createTaskModel"
	"issue-indexer/pckg/runtimeinfo"
	"issue-indexer/pckg/task"
)

func (service *TasksService) queueIsFree() bool {
	service.mx.Lock()
	defer service.mx.Unlock()
	if service.countNowRunnableTasks < service.MaxCountRunnableTasks {
		return true
	} else {
		return false
	}
}

func (service *TasksService) runTaskCompareIssuesInPairs(iTask task.ITask) error {
	taskState := iTask.(*Task)
	if service.queueIsFree() {
		comparable, compareWith, err := service.readIssuesForCompareInPairs(
			taskState.taskContext.(*createTaskModel.CreateTaskCompareIssuesInPairs).ComparableRepositoryID,
			taskState.taskContext.(*createTaskModel.CreateTaskCompareIssuesInPairs).CompareWithRepositoriesID,
		)
		if err != nil {
			runtimeinfo.LogError("DO NOT RUN TASK, READ FROM DB EXECUTE WITH ERROR: key: ", taskState.GetKey(), ", error: ", err)
			return err
		}
		//
		taskState.RunnableStatus = true
		//
		taskState.DeferStatus = false
		taskState.ExecutionStatus = false
		service.comparator.AddCompareIssuesInPairs(
			comparable,
			compareWith,
			taskState,
			service.comparator.CompareOnlyTitlesWithDictionaries,
		)
		runtimeinfo.LogInfo("TASK HAS BEEN SENT FOR EXECUTION: key: ", taskState.GetKey())
		service.countNowRunnableTasks++
	} else {
		taskState.DeferStatus = true
		//
		taskState.RunnableStatus = false
		taskState.ExecutionStatus = false
		taskState.runTaskTrigger = service.runTaskCompareIssuesInPairs
		runtimeinfo.LogInfo("TASK HAS BEEN SENT IN QUEUE: key: ", taskState.GetKey())
	}
	return nil
}
