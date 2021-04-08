package appService

import "errors"

var (
	ErrorQueueIsFilled = errors.New("Queue Is Filled. ")
)

//
//------------------------------------------CREATE TASK-----------------------------------------------------------------
//

type JsonRepository struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

type JsonCreateTaskRepositoriesDescriptions struct {
	TaskKey      string           `json:"task_key"`
	Repositories []JsonRepository `json:"repositories"`
}

type JsonCreateTaskRepositoryIssues struct {
	TaskKey    string         `json:"task_key"`
	Repository JsonRepository `json:"repository"`
}

type JsonCreateTaskRepositoriesByKeyWord struct {
	TaskKey string `json:"task_key"`
	KeyWord string `json:"key_word"`
}

//
//------------------------------------------UPDATE TASK-----------------------------------------------------------------
//

type JsonExecutionStatus struct {
	TaskKey       string `json:"task_key"`
	TaskCompleted bool   `json:"task_completed"`
}

type JsonUpdateTaskRepositoriesDescriptions struct {
	ExecutionTaskStatus JsonExecutionStatus   `json:"execution_task_status"`
	Repositories        []RepositoryDataModel `json:"repositories"`
}

type JsonUpdateTaskRepositoryIssues struct {
	ExecutionTaskStatus JsonExecutionStatus `json:"execution_task_status"`
	Issues              []IssueDataModel    `json:"issues"`
}

type JsonUpdateTaskRepositoriesByKeyWord struct {
	ExecutionTaskStatus JsonExecutionStatus   `json:"execution_task_status"`
	Repositories        []RepositoryDataModel `json:"repositories"`
}

//
//------------------------------------------DATA MODELS-----------------------------------------------------------------
//

type RepositoriesByKeyWordDataModel struct {
	Items []RepositoryDataModel `json:"items"`
}

type RepositoryDataModel struct {
	URL         string   `json:"url"`
	Topics      []string `json:"topics"`
	Description string   `json:"description"`
}

type ListIssuesDataModel []IssueDataModel

type IssueDataModel struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Body   string `json:"body"`
}
