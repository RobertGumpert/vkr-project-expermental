package main

type JSONExecutionTaskStatus struct {
	TaskKey       string `json:"task_key"`
	TaskCompleted bool   `json:"execution_status"`
}

type JSONRepository struct {
	URL    string   `json:"url"`
	Topics []string `json:"topics"`
	About  string   `json:"about"`
}

type JSONSendReposByURLS struct {
	ExecutionTaskStatus JSONExecutionTaskStatus `json:"execution_task_status"`
	Repositories        []JSONRepository        `json:"repositories"`
}
