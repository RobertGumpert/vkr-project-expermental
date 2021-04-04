package issueCompator

import "github.com/RobertGumpert/vkr-pckg/dataModel"

type ReturnResult func(result CompareResult)
type RuleForComparisonIssues func(a, b dataModel.IssueModel, comparisonSettings interface{}) (nearest dataModel.NearestIssuesModel, err error)
type RuleForSamplingComparableIssues func(samplingSettings interface{}) (toCompare, doNotCompare []dataModel.IssueModel, err error)
type RuleForNonComparableIssues func(doNotCompare []dataModel.IssueModel) (err error)

type CompareResult struct {
	identifier interface{}
	//
	nearestCompletedWithError []dataModel.NearestIssuesModel
	doNotCompare              []dataModel.IssueModel
	err                       error
}

func (c CompareResult) GetIdentifier() interface{} {
	return c.identifier
}

func (c CompareResult) GetNearestCompletedWithError() []dataModel.NearestIssuesModel {
	return c.nearestCompletedWithError
}

func (c CompareResult) GetDoNotCompare() []dataModel.IssueModel {
	return c.doNotCompare
}

func (c CompareResult) GetErr() error {
	return c.err
}

type compareRules struct {
	maxCountThreads int64
	//
	ruleForSamplingComparableIssues RuleForSamplingComparableIssues
	ruleForComparisonIssues         RuleForComparisonIssues
	returnResult                    ReturnResult
	ruleForNonComparableIssues      RuleForNonComparableIssues
	//
	comparisonSettings interface{}
	samplingSettings   interface{}
}

func newCompareRules(
	maxCountThreads int64,
	ruleForSamplingComparableIssues RuleForSamplingComparableIssues,
	ruleForComparisonIssues RuleForComparisonIssues,
	returnResult ReturnResult,
	ruleForNonComparableIssues RuleForNonComparableIssues,
	comparisonSettings interface{},
	samplingSettings interface{}) *compareRules {
	return &compareRules{maxCountThreads: maxCountThreads, ruleForSamplingComparableIssues: ruleForSamplingComparableIssues, ruleForComparisonIssues: ruleForComparisonIssues, returnResult: returnResult, ruleForNonComparableIssues: ruleForNonComparableIssues, comparisonSettings: comparisonSettings, samplingSettings: samplingSettings}
}

func (c compareRules) GetMaxCountThreads() int64 {
	return c.maxCountThreads
}

func (c compareRules) GetRuleForSamplingComparableIssues() RuleForSamplingComparableIssues {
	return c.ruleForSamplingComparableIssues
}

func (c compareRules) GetRuleForComparisonIssues() RuleForComparisonIssues {
	return c.ruleForComparisonIssues
}

func (c compareRules) GetReturnResult() ReturnResult {
	return c.returnResult
}

func (c compareRules) GetRuleForNonComparableIssues() RuleForNonComparableIssues {
	return c.ruleForNonComparableIssues
}

func (c compareRules) GetComparisonSettings() interface{} {
	return c.comparisonSettings
}

func (c compareRules) GetSamplingSettings() interface{} {
	return c.samplingSettings
}
