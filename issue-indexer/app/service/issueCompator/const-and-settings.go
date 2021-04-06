package issueCompator

import (
	"github.com/RobertGumpert/vkr-pckg/dataModel"
)

type ReturnResult func(result *CompareResult)
type RuleForComparisonIssues func(a, b dataModel.IssueModel, rules *CompareRules) (nearest dataModel.NearestIssuesModel, err error)
type RuleForSamplingComparableIssues func(rules *CompareRules) (toCompare, doNotCompare []dataModel.IssueModel, err error)

type countIntersectionForIssues map[uint]int64

type intersectionsForComparableRepository struct {
	Issues             map[uint]int64
	CountIntersections int64
}
