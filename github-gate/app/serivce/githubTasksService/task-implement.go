package githubTasksService

import "github-gate/pckg/task"

const (
	taskTypeRepositoriesDescriptionsByURL          task.Type = 0
	taskTypeRepositoryIssues                       task.Type = 1
	taskTypeRepositoriesDescriptionsAndTheirIssues task.Type = 2
)

type TaskForCollector struct {
	taskType        task.Type
	key             string
	executionStatus bool
	deferStatus     bool
	runnableStatus  bool
	result          interface{}
	//
	taskDetails *TaskDetails
}

func newTaskForCollector(taskType task.Type, key string, executionStatus bool, deferStatus bool, runnableStatus bool, result interface{}, taskDetails *TaskDetails) *TaskForCollector {
	return &TaskForCollector{
		taskType:        taskType,
		key:             key,
		executionStatus: executionStatus,
		deferStatus:     deferStatus,
		runnableStatus:  runnableStatus,
		result:          result,
		taskDetails:     taskDetails,
	}
}

func (t *TaskForCollector) SetType(taskType task.Type) {
	t.taskType = taskType
}

func (t *TaskForCollector) SetKey(taskKey string) {
	t.key = taskKey
}

func (t *TaskForCollector) SetExecutionStatus(flag bool) {
	t.executionStatus = flag
}

func (t *TaskForCollector) SetDeferStatus(flag bool) {
	t.deferStatus = flag
}

func (t *TaskForCollector) SetRunnableStatus(flag bool) {
	t.runnableStatus = flag
}

func (t *TaskForCollector) SetResult(result interface{}) {
	t.result = result
}

func (t *TaskForCollector) SetCustomFields(customFields interface{}) {
	return
}

func (t *TaskForCollector) GetType() task.Type {
	return t.taskType
}

func (t *TaskForCollector) GetKey() string {
	return t.key
}

func (t *TaskForCollector) GetExecutionStatus() bool {
	return t.executionStatus
}

func (t *TaskForCollector) GetDeferStatus() bool {
	return t.deferStatus
}

func (t *TaskForCollector) GetRunnableStatus() bool {
	return t.runnableStatus
}

func (t *TaskForCollector) GetResult() interface{} {
	return t.result
}

func (t *TaskForCollector) GetCustomFields() interface{} {
	return t.taskDetails
}
