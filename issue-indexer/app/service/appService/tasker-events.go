package appService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"issue-indexer/app/service/issueCompator"
)

func (service *AppService) eventRunTask(task itask.ITask) (doTaskAsDefer, deleteTask bool) {
	send := task.GetState().GetSendContext().(*sendContext)
	err := service.comparator.DOCompare(
		send.GetRules(),
		send.GetResult(),
	)
	if err != nil {
		task.GetState().SetError(err)
		//
		// RETURN RESULT
		//
		return true, false
	}
	return false, false
}

func (service *AppService) eventUpdateTaskState(task itask.ITask, interProgramUpdateContext interface{}) (err error) {
	update := interProgramUpdateContext.(*issueCompator.CompareResult)
	if update.GetErr() != nil {
		return update.GetErr()
	}
	task.GetState().SetCompleted(true)
	return nil
}

func (service *AppService) eventManageQueue(task itask.ITask) (deleteTasks, saveTasks map[string]struct{}) {
	deleteTasks, saveTasks = make(map[string]struct{}), make(map[string]struct{})
	switch task.GetType() {
	case compareWithGroupRepositories:
		deleteTasks[task.GetKey()] = struct{}{}
		break
	case compareBesideRepository:
		deleteTasks[task.GetKey()] = struct{}{}
		break
	}
	return deleteTasks, saveTasks
}
