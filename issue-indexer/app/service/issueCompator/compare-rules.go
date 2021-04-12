package issueCompator

type CompareRules struct {
	repositoryID uint
	//
	maxCountThreads int64
	//
	ruleForSamplingComparableIssues RuleForSamplingComparableIssues
	ruleForComparisonIssues         RuleForComparisonIssues
	returnResult                    ReturnResult
	//
	comparisonCondition interface{}
	samplingCondition   interface{}
}

func (c *CompareRules) GetSamplingCondition() interface{} {
	return c.samplingCondition
}

func (c *CompareRules) GetComparisonCondition() interface{} {
	return c.comparisonCondition
}

func (c *CompareRules) GetRepositoryID() uint {
	return c.repositoryID
}

func NewCompareRules(
	repositoryID uint,
	maxCountThreads int64,
	ruleForSamplingComparableIssues RuleForSamplingComparableIssues,
	ruleForComparisonIssues RuleForComparisonIssues,
	returnResult ReturnResult,
	comparisonSettings interface{},
	samplingSettings interface{}) *CompareRules {
	return &CompareRules{
		repositoryID:                    repositoryID,
		maxCountThreads:                 maxCountThreads,
		ruleForSamplingComparableIssues: ruleForSamplingComparableIssues,
		ruleForComparisonIssues:         ruleForComparisonIssues,
		returnResult:                    returnResult,
		comparisonCondition:             comparisonSettings,
		samplingCondition:               samplingSettings,
	}
}

func (c *CompareRules) GetMaxCountThreads() int64 {
	return c.maxCountThreads
}

func (c *CompareRules) GetRuleForSamplingComparableIssues() RuleForSamplingComparableIssues {
	return c.ruleForSamplingComparableIssues
}

func (c *CompareRules) GetRuleForComparisonIssues() RuleForComparisonIssues {
	return c.ruleForComparisonIssues
}

func (c *CompareRules) GetReturnResult() ReturnResult {
	return c.returnResult
}
