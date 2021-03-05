package updateTaskModel

//
//----------------------------------------------------------------------------------------------------------------------
//

type ExecutionTaskStatus struct {
	TaskKey       string `json:"task_key"`
	TaskCompleted bool   `json:"task_completed"`
}

//
//----------------------------------------------------------------------------------------------------------------------
//

type Repository struct {
	URL         string   `json:"url"`
	Topics      []string `json:"topics"`
	Description string   `json:"description"`
	//
	Err error `json:"err"`
}

type RepositoriesByURLS struct {
	ExecutionTaskStatus ExecutionTaskStatus `json:"execution_task_status"`
	Repositories        []Repository        `json:"repositories"`
}

//
//----------------------------------------------------------------------------------------------------------------------
//


type Issue struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Body   string `json:"body"`
	//
	Err error `json:"err"`
}

type RepositoryIssues struct {
	ExecutionTaskStatus ExecutionTaskStatus `json:"execution_task_status"`
	Issues              []Issue             `json:"issues"`
}
