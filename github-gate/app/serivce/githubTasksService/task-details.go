package githubTasksService

type TaskDetails struct {
	number int
	//
	signalTriggeredDependentTask     bool
	dependentTasksRunAfterCompletion []*TaskForCollector
	//
	sendToCollectorJsonBody interface{}
	collectorAddress        string
	collectorEndpoint       string
	collectorURL            string
}


//
//
//

func (t *TaskDetails) SetDependentStatus(flag bool) {
	t.signalTriggeredDependentTask = flag
}

func (t *TaskDetails) IsDependent() bool {
	return t.signalTriggeredDependentTask
}

//
//
//

func (t *TaskDetails) SetDependentTasks(depend []*TaskForCollector) {
	t.dependentTasksRunAfterCompletion = depend
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