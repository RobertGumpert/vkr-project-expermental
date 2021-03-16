package githubCollectorModels

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

type UpdateTaskRepository struct {
	URL         string   `json:"url"`
	Topics      []string `json:"topics"`
	Description string   `json:"description"`
	//
	Err error `json:"err"`
}

type UpdateTaskRepositoriesByURLS struct {
	ExecutionTaskStatus ExecutionTaskStatus    `json:"execution_task_status"`
	Repositories        []UpdateTaskRepository `json:"repositories"`
}

//
//----------------------------------------------------------------------------------------------------------------------
//


type UpdateTaskIssue struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Body   string `json:"body"`
	//
	Err error `json:"err"`
}

type UpdateTaskRepositoryIssues struct {
	ExecutionTaskStatus ExecutionTaskStatus `json:"execution_task_status"`
	Issues              []UpdateTaskIssue   `json:"issues"`
}
