package issueCompator

import (
	"errors"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"runtime"
	"sync"
)

type Comparator struct {
	db repository.IRepository
}

func (comparator *Comparator) CreateCompareResult(
	maxCountThreads int64,
	ruleForSamplingComparableIssues RuleForSamplingComparableIssues,
	ruleForComparisonIssues RuleForComparisonIssues,
	returnResult ReturnResult,
	ruleForNonComparableIssues RuleForNonComparableIssues,
	comparisonSettings interface{},
	samplingSettings interface{}) compareRules {
	return compareRules{
		maxCountThreads:                 maxCountThreads,
		ruleForSamplingComparableIssues: ruleForSamplingComparableIssues,
		ruleForComparisonIssues:         ruleForComparisonIssues,
		returnResult:                    returnResult,
		ruleForNonComparableIssues:      ruleForNonComparableIssues,
		comparisonSettings:              comparisonSettings,
		samplingSettings:                samplingSettings,
	}
}

func (comparator *Comparator) NewOneWithAll(repositoryID uint, countThreads int, identifier interface{}, rules compareRules) (result CompareResult, err error) {
	if repositoryID == 0 {
		return result, errors.New("RepositoryID is 0. ")
	}
	whatToCompare, err := comparator.db.GetIssueRepository(repositoryID)
	if err != nil {
		return result, err
	}
	if len(whatToCompare) == 0 {
		return result, errors.New("Size of slice repository issues is 0. ")
	}
	result = CompareResult{
		identifier:                identifier,
		nearestCompletedWithError: make([]dataModel.NearestIssuesModel, 0),
		doNotCompare:              make([]dataModel.IssueModel, 0),
		err:                       nil,
	}
	go func(comparator *Comparator, rules compareRules, result CompareResult, whatToCompare []dataModel.IssueModel) {
		comparator.doCompareIntoMultipleStreams(
			rules,
			result,
			whatToCompare,
		)
		return
	}(comparator, rules, result, whatToCompare)
	return result, nil
}

func (comparator *Comparator) iterating(whatToCompare, whatToCompareWith []dataModel.IssueModel, from, to int64, rules compareRules, result CompareResult, wg *sync.WaitGroup) {
	for i := from; i < to; i++ {
		for j := 0; j < len(whatToCompareWith); j++ {
			a := whatToCompare[i]
			b := whatToCompareWith[j]
			nearest, err := rules.ruleForComparisonIssues(
				a,
				b,
				rules.GetComparisonSettings(),
			)
			if err != nil {
				continue
			}
			err = comparator.db.AddNearestIssues(nearest)
			if err != nil {
				continue
			} else {
				result.nearestCompletedWithError = append(
					result.nearestCompletedWithError,
					nearest,
				)
			}
		}
	}
	runtime.GC()
	if wg != nil {
		wg.Done()
	}
	return
}

func (comparator *Comparator) doCompareIntoMultipleStreams(rules compareRules, result CompareResult, whatToCompare []dataModel.IssueModel) {
	var (
		whatToCompareWith, doNotCompare   []dataModel.IssueModel
		err                               error
		lengthOfPartComparableIssuesSlice int64
		wg                                = new(sync.WaitGroup)
		from, to                          int64
	)
	whatToCompareWith, doNotCompare, err = rules.GetRuleForSamplingComparableIssues()(rules.GetSamplingSettings())
	if err != nil {
		return
	}
	if len(whatToCompareWith) == 0 {
		return
	}
	result.doNotCompare = doNotCompare
	lengthOfPartComparableIssuesSlice = int64(len(whatToCompareWith)) / rules.GetMaxCountThreads()
	if lengthOfPartComparableIssuesSlice <= 1 {
		comparator.iterating(
			whatToCompare,
			whatToCompareWith,
			int64(0),
			int64(len(whatToCompare)),
			rules,
			result,
			nil,
		)
	} else {
		to = lengthOfPartComparableIssuesSlice
		for {
			if to >= int64(len(whatToCompare)) {
				wg.Add(1)
				go comparator.iterating(
					whatToCompare,
					whatToCompareWith,
					from,
					int64(len(whatToCompare)),
					rules,
					result,
					wg,
				)
				break
			}
			wg.Add(1)
			go comparator.iterating(
				whatToCompare,
				whatToCompareWith,
				from,
				to,
				rules,
				result,
				wg,
			)
			from = to + 1
			to = from + lengthOfPartComparableIssuesSlice
		}
		wg.Wait()
	}
	rules.GetReturnResult()(result)
	return
}
