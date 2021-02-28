package main

type UpdateTaskStateExecutionStatus struct {
	TaskKey       string `json:"task_key"`
	TaskCompleted bool   `json:"task_completed"`
}

//
//--------------TASK GET REPOSITORIES BY URLS---------------------------------------------------------------------------
//

type UpdateTaskStateRepository struct {
	URL         string   `json:"url"`
	Topics      []string `json:"topics"`
	Description string   `json:"description"`
	//
	Err error `json:"err"`
}

type UpdateTaskStateReposByURLS struct {
	ExecutionTaskStatus UpdateTaskStateExecutionStatus `json:"execution_task_status"`
	Repositories        []UpdateTaskStateRepository    `json:"repositories"`
}

//
//--------------TASK GET REPOSITORY ISSUE-------------------------------------------------------------------------------
//

type UpdateTaskStateIssue struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	State  string   `json:"state"`
	Body   string `json:"body"`
	//
	Err error `json:"err"`
}

type UpdateTaskStateRepositoryIssues struct {
	ExecutionTaskStatus UpdateTaskStateExecutionStatus `json:"execution_task_status"`
	Issues              []UpdateTaskStateIssue         `json:"issues"`
}
