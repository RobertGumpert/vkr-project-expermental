package githubCollectorService

import (
	"errors"
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/itask"
)

const (
	gitHubApiAddress                          = "https://api.github.com"
	collectorEndpointRepositoriesDescriptions = "api/task/repositories/descriptions"
	collectorEndpointRepositoryIssues         = "api/task/repository/issues"
	collectorEndpointRepositoriesByKeyWord    = "api/task/repositories/by/keyword"
)

const (
	TaskTypeDownloadOnlyDescriptions                                    itask.Type = 1000
	TaskTypeDownloadOnlyIssues                                          itask.Type = 1001
	TaskTypeDownloadCompositeByName                                     itask.Type = 1002
	TaskTypeDownloadCompositeByKeyWord                                  itask.Type = 1003
	TaskTypeDownloadCompositeRepositoryAndRepositoriesContainingKeyWord itask.Type = 1004
)

//
// COMPOSITE CUSTOM FIELDS----------------------------------------------------------------------------------------------
//

type compositeCustomFields struct {
	TaskType itask.Type
	Fields   interface{}
}

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

type jsonSendToCollectorRepository struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

type jsonSendToCollectorDescriptionsRepositories struct {
	TaskKey      string                          `json:"task_key"`
	Repositories []jsonSendToCollectorRepository `json:"repositories"`
}

type jsonSendToCollectorRepositoryIssues struct {
	TaskKey    string                        `json:"task_key"`
	Repository jsonSendToCollectorRepository `json:"repository"`
}

type jsonSendToCollectorRepositoriesByKeyWord struct {
	TaskKey string `json:"task_key"`
	KeyWord string `json:"key_word"`
}

//
// Models:
//

type jsonSendFromCollectorRepository struct {
	URL         string   `json:"url"`
	Topics      []string `json:"topics"`
	Description string   `json:"description"`
	Err         error    `json:"err"`
}

type jsonSendFromCollectorIssue struct {
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
	ExecutionTaskStatus jsonExecutionTaskStatus           `json:"execution_task_status"`
	Repositories        []jsonSendFromCollectorRepository `json:"repositories"`
}

type jsonSendFromCollectorRepositoryIssues struct {
	ExecutionTaskStatus jsonExecutionTaskStatus      `json:"execution_task_status"`
	Issues              []jsonSendFromCollectorIssue `json:"issues"`
}

type jsonSendFromCollectorRepositoriesByKeyWord struct {
	ExecutionTaskStatus jsonExecutionTaskStatus           `json:"execution_task_status"`
	Repositories        []jsonSendFromCollectorRepository `json:"repositories"`
}

//
// ERROR----------------------------------------------------------------------------------------------------------------
//

var (
	ErrorTaskTypeNotExist  = errors.New("Task Type Not Exist. ")
	ErrorNoFreeCollector   = errors.New("No Free Collector. ")
	ErrorCollectorIsBusy   = errors.New("Collector Is Busy. ")
	ErrorNoneCorrectData   = errors.New("Not Full Send Context. ")
	ErrorTaskIsNilPointer  = errors.New("Task is nil pointer. ")
	ErrorRepositoryIsExist = errors.New("Repository is exist. ")
	ErrorQueueIsFilled     = gotasker.ErrorQueueIsFilled
)
