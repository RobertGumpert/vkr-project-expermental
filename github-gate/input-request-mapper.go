package main

//
//---------------------CREATE TASK--------------------------------------------------------------------------------------
//

type CreateTaskRepoByURLS struct {
	Repositories []string `json:"repositories"`
}

//
//---------------------UPDATE TASK--------------------------------------------------------------------------------------
//

type UpdateTaskExecutionTaskStatus struct {
	TaskKey       string `json:"task_key"`
	TaskCompleted bool   `json:"task_completed"`
}

type UpdateTaskRepository struct {
	URL         string   `json:"url"`
	Topics      []string `json:"topics"`
	Description string   `json:"description"`
	//
	Err error `json:"err"`
}

type UpdateTaskReposByURLS struct {
	ExecutionTaskStatus UpdateTaskExecutionTaskStatus `json:"execution_task_status"`
	Repositories        []UpdateTaskRepository        `json:"repositories"`
}
