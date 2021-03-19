package githubTasksService

import "github-gate/pckg/task"

type TaskDetails struct {
	taskFromTaskService task.ITask
	number              int
	//
	entityID uint
	//
	dependentStatus bool
	triggerTask     *TaskForCollector
	//
	triggeredStatus bool
	dependentTasks  []*TaskForCollector
	//
	sendToCollectorJsonBody interface{}
	collectorAddress        string
	collectorEndpoint       string
	collectorURL            string
}

//
//
//

func (t *TaskDetails) GetTaskFromTaskService() task.ITask {
	return t.taskFromTaskService
}

func (t *TaskDetails) SetTaskFromTaskService(taskFromTaskService task.ITask) {
	t.taskFromTaskService = taskFromTaskService
}

//
//
//

func (t *TaskDetails) GetEntityID() uint {
	return t.entityID
}

func (t *TaskDetails) SetEntityID(entityID uint) {
	t.entityID = entityID
}

//
//
//

func (t *TaskDetails) GetTriggerTask() *TaskForCollector {
	return t.triggerTask
}

func (t *TaskDetails) SetTriggerTask(triggerTask *TaskForCollector) {
	t.triggerTask = triggerTask
}

func (t *TaskDetails) CountCompletedDependentTasks() int {
	var count int
	for i := 0; i < len(t.dependentTasks); i++ {
		if t.dependentTasks[i].GetExecutionStatus() {
			count++
		}
	}
	return count
}

//
//
//

func (t *TaskDetails) SetTriggeredStatus(flag bool) {
	t.triggeredStatus = flag
}

func (t *TaskDetails) IsTrigger() bool {
	return t.triggeredStatus
}

func (t *TaskDetails) SetDependentStatus(flag bool) {
	t.dependentStatus = flag
}

func (t *TaskDetails) IsDependent() bool {
	return t.dependentStatus
}

//
//
//

func (t *TaskDetails) SetDependentTasks(depend []*TaskForCollector) {
	t.dependentTasks = depend
}

func (t *TaskDetails) HasDependentTasks() (bool, []*TaskForCollector) {
	var isExistDependentTasks bool
	if t.dependentTasks == nil || len(t.dependentTasks) == 0 {
		isExistDependentTasks = false
	} else {
		isExistDependentTasks = true
	}
	return isExistDependentTasks, t.dependentTasks
}

//
//
//

func (t *TaskDetails) SetNumber(number int) {
	t.number = number
}

func (t *TaskDetails) GetNumber() int {
	return t.number
}

//
//
//

func (t *TaskDetails) SetCollectorAddress(collectorAddress string) {
	t.collectorAddress = collectorAddress
}

func (t *TaskDetails) GetCollectorAddress() string {
	return t.collectorAddress
}

//
//
//

func (t *TaskDetails) SetCollectorURL(collectorURL string) {
	t.collectorURL = collectorURL
}

func (t *TaskDetails) GetCollectorURL() string {
	return t.collectorURL
}

//
//
//

func (t *TaskDetails) SetSendToCollectorJsonBody(sendToCollectorJsonBody interface{}) {
	t.sendToCollectorJsonBody = sendToCollectorJsonBody
}

func (t *TaskDetails) GetSendToCollectorJsonBody() interface{} {
	return t.sendToCollectorJsonBody
}

//
//
//

func (t *TaskDetails) GetCollectorEndpoint() string {
	return t.collectorEndpoint
}

func (t *TaskDetails) SetCollectorEndpoint(collectorEndpoint string) {
	t.collectorEndpoint = collectorEndpoint
}
