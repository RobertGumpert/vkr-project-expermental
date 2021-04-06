package appService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"issue-indexer/app/service/issueCompator"
)

const (
	compareWithGroupRepositories itask.Type = 0
	compareBesideRepository      itask.Type = 1
)

type sendContext struct {
	rules  *issueCompator.CompareRules
	result *issueCompator.CompareResult
}

func (s *sendContext) GetResult() *issueCompator.CompareResult {
	return s.result
}

func (s *sendContext) GetRules() *issueCompator.CompareRules {
	return s.rules
}
