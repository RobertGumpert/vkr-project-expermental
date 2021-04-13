package appService

import "github.com/RobertGumpert/gotasker/itask"

const (
	taskTypeReindexingForAll               itask.Type = 0
	taskTypeReindexingForRepository        itask.Type = 1
	taskTypeReindexingForGroupRepositories itask.Type = 2
)

//
//
//

type resultIndexing struct {
	taskType itask.Type
	taskKey  string
	jsonBody interface{}
}
type doReindexing func()
type setReindexingContext func(settings interface{})

//
//------------------------------------------------------JSON------------------------------------------------------------
//

type jsonNearestRepository struct {
	RepositoryID          uint             `json:"repository_id"`
	NearestRepositoriesID map[uint]float64 `json:"nearest_repositories_id"`
}

type jsonExecutionTaskStatus struct {
	TaskKey       string `json:"task_key"`
	TaskCompleted bool   `json:"task_completed"`
	Error         error  `json:"error"`
}

type jsonSendFromGateReindexingForRepository struct {
	TaskKey      string `json:"task_key"`
	RepositoryID uint   `json:"repository_id"`
}

type jsonSendFromGateReindexingForAll struct {
	TaskKey string `json:"task_key"`
}

type jsonSendFromGateReindexingForGroupRepositories struct {
	TaskKey      string `json:"task_key"`
	RepositoryID []uint `json:"repository_id"`
}

type jsonSendToGateReindexingForRepository struct {
	ExecutionTaskStatus jsonExecutionTaskStatus `json:"execution_task_status"`
	Result              jsonNearestRepository   `json:"result"`
}

type jsonSendToGateReindexingForAll struct {
	ExecutionTaskStatus jsonExecutionTaskStatus `json:"execution_task_status"`
	Results             []jsonNearestRepository `json:"results"`
}

type jsonSendToGateReindexingForGroupRepositories struct {
	ExecutionTaskStatus jsonExecutionTaskStatus `json:"execution_task_status"`
	Results             []jsonNearestRepository `json:"results"`
}

//
//------------------------------------------------------JSON------------------------------------------------------------
//

type jsonInputWordIsExist struct {
	Word string `json:"word"`
}

type jsonInputNearestRepositoriesForRepository struct {
	RepositoryID uint `json:"repository_id"`
}

type jsonOutputWordIsExist struct {
	WordIsExist          bool `json:"word_is_exist"`
	DatabaseIsReindexing bool `json:"database_is_reindexing"`
}

type jsonOutputNearestRepositoriesForRepository struct {
	NearestRepositories  []jsonNearestRepository `json:"nearest_repositories"`
	DatabaseIsReindexing bool                    `json:"database_is_reindexing"`
}
