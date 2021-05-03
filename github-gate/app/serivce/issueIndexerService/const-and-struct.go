package issueIndexerService

import "github.com/RobertGumpert/gotasker/itask"

const(
	TaskTypeCompareIssuesGroupRepositories itask.Type = 2000
)

//
//
//

type JsonSendToIndexerCompareGroup struct {
	TaskKey                  string `json:"task_key"`
	RepositoryID             uint   `json:"repository_id"`
	ComparableRepositoriesID []uint `json:"comparable_repositories_id"`
}

type JsonExecutionTaskStatus struct {
	TaskKey       string `json:"task_key"`
	TaskCompleted bool   `json:"task_completed"`
	Error         error  `json:"error"`
}

type JsonSendFromIndexerCompareGroup struct {
	ExecutionTaskStatus JsonExecutionTaskStatus `json:"execution_task_status"`
}