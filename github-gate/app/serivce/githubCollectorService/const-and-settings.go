package githubCollectorService

import (
	"errors"
	"github.com/RobertGumpert/gotasker/itask"
)

const (
	RepositoriesDescription          itask.Type = 0
	RepositoryIssues                 itask.Type = 1
	RepositoriesDescriptionAndIssues itask.Type = 2
)

//
// CONTEXT--------------------------------------------------------------------------------------------------------------
//

type contextTaskSend struct {
	CollectorAddress, CollectorEndpoint, CollectorURL string
	JSONBody                                          interface{}
}

//
// JSON-----------------------------------------------------------------------------------------------------------------
//

//
// Send:
//

type jsonSendToCollectorDescriptionsRepositories struct {
	TaskKey string   `json:"task_key"`
	URLS    []string `json:"urls"`
}

type jsonSendToCollectorRepositoryIssues struct {
	TaskKey string `json:"task_key"`
	URL     string `json:"url"`
}

//
// Models:
//

type repositoryDescription struct {
	URL         string   `json:"url"`
	Topics      []string `json:"topics"`
	Description string   `json:"description"`
	Err         error    `json:"err"`
}

type issueDescription struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Body   string `json:"body"`
	Err    error  `json:"err"`
}

//
// From:
//

type jsonExecutionTaskStatus struct {
	TaskKey       string `json:"task_key"`
	TaskCompleted bool   `json:"task_completed"`
}

type jsonSendFromCollectorDescriptionsRepositories struct {
	ExecutionTaskStatus jsonExecutionTaskStatus `json:"execution_task_status"`
	Repositories        []repositoryDescription `json:"repositories"`
}

type jsonSendFromCollectorRepositoryIssues struct {
	ExecutionTaskStatus jsonExecutionTaskStatus `json:"execution_task_status"`
	Issues              []issueDescription      `json:"issues"`
}

//
// ERROR----------------------------------------------------------------------------------------------------------------
//

var (
	ErrorTaskTypeNotExist   = errors.New("Task Type Not Exist. ")
	ErrorNoFreeCollector    = errors.New("No Free Collector. ")
	ErrorCollectorIsBusy    = errors.New("Collector Is Busy. ")
	ErrorNotFullSendContext = errors.New("Not Full Send Context. ")
	ErrorTaskIsNilPointer   = errors.New("Task is nil pointer. ")
)
