package appService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"issue-indexer/app/service/issueCompator"
)

const (
	compareWithGroupRepositories itask.Type = 0
	compareBesideRepository      itask.Type = 1
)

//
//
//

type sendToGateContext struct {
	endpoint string
	taskKey  string
	err      error
	jsonBody interface{}
}

func (s *sendToGateContext) GetErr() error {
	return s.err
}

func (s *sendToGateContext) GetJsonBody() interface{} {
	return s.jsonBody
}

func (s *sendToGateContext) GetTaskKey() string {
	return s.taskKey
}

func (s *sendToGateContext) GetEndpoint() string {
	return s.endpoint
}

//
//
//

type sendToComparatorContext struct {
	rules  *issueCompator.CompareRules
	result *issueCompator.CompareResult
}

func (s *sendToComparatorContext) GetResult() *issueCompator.CompareResult {
	return s.result
}

func (s *sendToComparatorContext) GetRules() *issueCompator.CompareRules {
	return s.rules
}

//
//--------------------------------------------------JSON----------------------------------------------------------------
//

type jsonExecutionTaskStatus struct {
	TaskKey       string `json:"task_key"`
	TaskCompleted bool   `json:"task_completed"`
	Error         error  `json:"error"`
}

type jsonSendToGateCompareGroup struct {
	ExecutionTaskStatus jsonExecutionTaskStatus `json:"execution_task_status"`
}

type jsonSendToGateCompareBeside struct {
	ExecutionTaskStatus jsonExecutionTaskStatus `json:"execution_task_status"`
}

type jsonSendFromGateCompareGroup struct {
	TaskKey                  string `json:"task_key"`
	RepositoryID             uint   `json:"repository_id"`
	ComparableRepositoriesID []uint `json:"comparable_repositories_id"`
}

type jsonSendFromCompareBeside struct {
	TaskKey      string `json:"task_key"`
	RepositoryID uint   `json:"repository_id"`
}
