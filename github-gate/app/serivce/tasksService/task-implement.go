package tasksService

import "github-gate/pckg/task"

const(
	taskTypeAddRepositories = 0
)

type Task struct {
	Type            task.Type
	Key             string
	ExecutionStatus bool
	DeferStatus     bool
	RunnableStatus  bool
	Result          interface{}
	//
	TaskContext interface{}
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
	t.Result = result
}

func (t *Task) SetCustomFields(customFields interface{}) {
	return
}

func (t *Task) GetType() task.Type {
	return t.Type
}

func (t *Task) GetKey() string {
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
	return t.Result
}

func (t *Task) GetCustomFields() interface{} {
	return t.TaskContext
}
