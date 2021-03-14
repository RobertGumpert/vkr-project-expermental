package tasksService

import "issue-indexer/pckg/task"

const (
	TypeTaskCompareIssuesInPairs task.Type = 0
)

type Task struct {
	Type                        task.Type
	Key                         string
	ExecutionStatus             bool
	RunnableStatus              bool
	ResultCompareFromComparator interface{}
	DeferStatus                 bool
	//
	taskContext      interface{}
	sendResultToGate interface{}
	runTaskTrigger   func(iTask task.ITask) error
}

func (t *Task) SetType(taskType task.Type) {
	t.Type = taskType
}

func (t *Task) SetKey(taskKey string) {
	t.Key = taskKey
}

func (t *Task) SetExecutionStatus(flag bool) {
	t.ExecutionStatus = flag
}

func (t *Task) SetDeferStatus(flag bool) {
	t.DeferStatus = flag
}

func (t *Task) SetRunnableStatus(flag bool) {
	t.RunnableStatus = flag
}

func (t *Task) SetResult(result interface{}) {
	t.ResultCompareFromComparator = result
}

func (t *Task) SetCustomFields(customFields interface{}) {
	return
}

func (t *Task) GetType() task.Type {
	return t.Type
}

func (t Task) GetKey() string {
	return t.Key
}

func (t *Task) GetExecutionStatus() bool {
	return t.ExecutionStatus
}

func (t *Task) GetDeferStatus() bool {
	return t.DeferStatus
}

func (t *Task) GetRunnableStatus() bool {
	return t.RunnableStatus
}

func (t *Task) GetResult() interface{} {
	return t.ResultCompareFromComparator
}

func (t *Task) GetCustomFields() interface{} {
	return t.taskContext
}
