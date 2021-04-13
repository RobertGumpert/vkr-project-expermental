package taskManager

import (
	"github.com/RobertGumpert/gotasker/itask"
)

type TaskManager struct {
	sliceTasks []itask.ITask
}

func (manager *TaskManager) CreateTask(t itask.Type, key string, send, update, fields interface{}, eventRunTask itask.EventRunTask, eventUpdateState itask.EventUpdateTaskState) (task itask.ITask, err error) {
	panic("implement me")
}

func (manager *TaskManager) ModifyTaskAsTrigger(trigger itask.ITask, dependents ...itask.ITask) (task itask.ITask, err error) {
	panic("implement me")
}

func (manager *TaskManager) RunTask(task itask.ITask) (doTaskAsDefer, sendToErrorChannel bool) {
	panic("implement me")
}

func (manager *TaskManager) AddTaskAndTask(task itask.ITask) (err error) {
	panic("implement me")
}

func (manager *TaskManager) RunDependentTasks(task itask.ITask) {
	panic("implement me")
}

func (manager *TaskManager) RunDeferTasks(runDependentTasks bool) {
	panic("implement me")
}

func (manager *TaskManager) RunDeferByTimer() {
	panic("implement me")
}

func (manager *TaskManager) CreateError(e error, taskKey string, task itask.ITask) (err itask.IError) {
	panic("implement me")
}

func (manager *TaskManager) SendErrorToErrorChannel(err itask.IError) {
	panic("implement me")
}

func (manager *TaskManager) SetUpdateForTask(key string, somethingUpdateContext interface{}) {
	panic("implement me")
}

func (manager *TaskManager) ManageUpdates() {
	panic("implement me")
}

func (manager *TaskManager) ManageCompleted() {
	panic("implement me")
}

func (manager *TaskManager) GetChannelError() (channelForSendErrors chan itask.IError) {
	panic("implement me")
}

func (manager *TaskManager) GetSizeQueue() (sizeOfQueue int64) {
	panic("implement me")
}

func (manager *TaskManager) QueueIsFilled(countTasks int64) (isFilled bool) {
	panic("implement me")
}

func (manager *TaskManager) DeleteTasksByKeys(keys map[string]struct{}) {
	panic("implement me")
}

func (manager *TaskManager) FindTaskByKey(key string) (findTask itask.ITask, err error) {
	panic("implement me")
}

func (manager *TaskManager) FindRunBanTriggers() (runBanTriggers []itask.ITask) {
	panic("implement me")
}

func (manager *TaskManager) FindRunBanSimpleTasks() (runBanTasks []itask.ITask) {
	panic("implement me")
}

func (manager *TaskManager) FindDependentTasksIfTriggerNotExist(triggerKey string) (dependentsTasks []itask.ITask) {
	panic("implement me")
}

func (manager *TaskManager) SetRunBan(tasks ...itask.ITask) {
	panic("implement me")
}

func (manager *TaskManager) SetRunBanInQueue(tasks ...itask.ITask) {
	panic("implement me")
}

func (manager *TaskManager) TakeOffRunBanInQueue(tasks ...itask.ITask) {
	panic("implement me")
}

func (manager *TaskManager) TriggerIsCompleted(trigger itask.ITask) (isCompleted bool, dependentTasks map[string]bool, err error) {
	panic("implement me")
}

