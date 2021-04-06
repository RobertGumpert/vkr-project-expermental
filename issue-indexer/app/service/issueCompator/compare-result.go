package issueCompator

import "github.com/RobertGumpert/vkr-pckg/dataModel"

type CompareResult struct {
	identifier                interface{}
	nearestCompletedWithError []dataModel.NearestIssuesModel
	doNotCompare              []dataModel.IssueModel
	err                       error
}

func NewCompareResult(identifier interface{}) *CompareResult {
	return &CompareResult{
		identifier:                identifier,
		nearestCompletedWithError: make([]dataModel.NearestIssuesModel, 0),
		doNotCompare:              make([]dataModel.IssueModel, 0),
		err:                       nil,
	}
}

func (c *CompareResult) GetIdentifier() interface{} {
	return c.identifier
}

func (c *CompareResult) GetNearestCompletedWithError() []dataModel.NearestIssuesModel {
	return c.nearestCompletedWithError
}

func (c *CompareResult) GetDoNotCompare() []dataModel.IssueModel {
	return c.doNotCompare
}

func (c *CompareResult) GetErr() error {
	return c.err
}
