package repositoryIndexerService

import "github.com/RobertGumpert/gotasker/itask"

const(
	TaskTypeReindexingForRepository itask.Type = 3000
)


//
//
//


type JsonNearestRepository struct {
	RepositoryID          uint             `json:"repository_id"`
	NearestRepositoriesID map[uint]float64 `json:"nearest_repositories_id"`
}

type JsonExecutionTaskStatus struct {
	TaskKey       string `json:"task_key"`
	TaskCompleted bool   `json:"task_completed"`
	Error         error  `json:"error"`
}


//
//
//


type JsonSendToIndexerReindexingForRepository struct {
	TaskKey      string `json:"task_key"`
	RepositoryID uint   `json:"repository_id"`
}

type JsonSendFromIndexerReindexingForRepository struct {
	ExecutionTaskStatus JsonExecutionTaskStatus `json:"execution_task_status"`
	Result              JsonNearestRepository   `json:"result"`
}

