package githubTasksService

type TaskDetails struct {
	signalTriggeredDependentTask     bool
	dependentTasksRunAfterCompletion []*TaskForCollector
	//
	sendToCollectorJsonBody interface{}
	collectorAddress        string
	collectorURL            string
}

func NewTaskDetails(signalTriggeredDependentTask bool, relatedTasks []*TaskForCollector, sendToCollectorJsonBody interface{}, collectorAddress string, collectorURL string) *TaskDetails {
	return &TaskDetails{
		signalTriggeredDependentTask:     signalTriggeredDependentTask,
		dependentTasksRunAfterCompletion: relatedTasks,
		sendToCollectorJsonBody:          sendToCollectorJsonBody,
		collectorAddress:                 collectorAddress,
		collectorURL:                     collectorURL,
	}
}

func (t *TaskDetails) IsDependent() bool {
	return t.signalTriggeredDependentTask
}

func (t *TaskDetails) HasDependentTasks() (bool, []*TaskForCollector) {
	var isExistDependentTasks bool
	if t.dependentTasksRunAfterCompletion == nil || len(t.dependentTasksRunAfterCompletion) == 0 {
		isExistDependentTasks = false
	} else {
		isExistDependentTasks = true
	}
	return isExistDependentTasks, t.dependentTasksRunAfterCompletion
}

func (t *TaskDetails) GetCollectorAddress() string {
	return t.collectorAddress
}

func (t *TaskDetails) GetCollectorURL() string {
	return t.collectorURL
}

func (t *TaskDetails) GetSendToCollectorJsonBody() interface{} {
	return t.sendToCollectorJsonBody
}

