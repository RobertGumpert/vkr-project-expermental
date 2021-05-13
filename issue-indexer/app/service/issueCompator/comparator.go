package issueCompator

import (
	"errors"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"gorm.io/gorm"
	"runtime"
	"sync"
)

type Comparator struct {
	db repository.IRepository
	mx *sync.Mutex
}

func NewComparator(db repository.IRepository) *Comparator {
	return &Comparator{
		db: db,
		mx: new(sync.Mutex),
	}
}

func (comparator *Comparator) DOCompare(rules *CompareRules, result *CompareResult) (err error) {
	if rules.GetRepositoryID() == 0 {
		return errors.New("RepositoryID is 0. ")
	}
	whatToCompare, err := comparator.db.GetIssueRepository(rules.GetRepositoryID())
	if err != nil {
		return err
	}
	if len(whatToCompare) == 0 {
		return errors.New("Size of slice repository issues is 0. ")
	}
	go func(comparator *Comparator, rules *CompareRules, result *CompareResult, whatToCompare []dataModel.IssueModel) {
		comparator.doCompareIntoMultipleStreams(
			rules,
			result,
			whatToCompare,
		)
		return
	}(comparator, rules, result, whatToCompare)
	return nil
}

func (comparator *Comparator) iterating(whatToCompare, whatToCompareWith []dataModel.IssueModel, from, to int64, rules *CompareRules, result *CompareResult, intersections map[uint]countIntersectionForIssues, wg *sync.WaitGroup) {
	var (
		nearestIssues = make([]dataModel.NearestIssuesModel, 0)
	)
	for i := from; i < to; i++ {
		for j := 0; j < len(whatToCompareWith); j++ {
			a := whatToCompare[i]
			b := whatToCompareWith[j]
			nearest, isNearest := rules.ruleForComparisonIssues(
				a,
				b,
				rules,
			)
			if isNearest == nil {
				nearestIssues = append(nearestIssues, nearest)
			}
			comparator.mx.Lock()
			intersection, repositoryIsExist := intersections[b.RepositoryID]
			if !repositoryIsExist {
				intersection = make(map[uint]int64)
				intersections[b.RepositoryID] = intersection
			}
			if _, issueIsExist := intersection[b.ID]; !issueIsExist {
				intersection[b.ID] = 0
			}
			if isNearest == nil {
				intersection[b.ID] = intersection[b.ID] + 1
			}
			comparator.mx.Unlock()
		}
	}
	if len(nearestIssues) != 0 {
		err := comparator.db.AddListNearestIssues(nearestIssues)
		if err != nil {
			runtimeinfo.LogError(err)
			for _, nearest := range nearestIssues {
				err := comparator.db.AddNearestIssues(nearest)
				if err != nil {
					runtimeinfo.LogError(err)
				}
			}
		}
	}
	runtime.GC()
	if wg != nil {
		wg.Done()
	}
	return
}

func (comparator *Comparator) doCompareIntoMultipleStreams(rules *CompareRules, result *CompareResult, repositoryIssues []dataModel.IssueModel) {
	var (
		comparableIssues, doNotCompare    []dataModel.IssueModel
		err                               error
		lengthOfPartComparableIssuesSlice int64
		from, to                          int64
		//
		taskKey               = result.identifier.(itask.ITask).GetKey()
		intersectionModels    = make([]dataModel.NumberIssueIntersectionsModel, 0)
		repositoryID          = rules.GetRepositoryID()
		countIssuesRepository = int64(len(repositoryIssues))
		intersections         = make(map[uint]countIntersectionForIssues)
		wg                    = new(sync.WaitGroup)
	)
	runtimeinfo.LogInfo("START COMPARE TASK [", taskKey, "].")
	comparableIssues, doNotCompare, err = rules.GetRuleForSamplingComparableIssues()(rules)
	if err != nil {
		result.err = err
		rules.GetReturnResult()(result)
		return
	}
	if len(comparableIssues) == 0 {
		result.err = errors.New("LIST WHAT TO COMPARE IS EMPTY. ")
		rules.GetReturnResult()(result)
		return
	}
	result.doNotCompare = doNotCompare
	lengthOfPartComparableIssuesSlice = int64(len(repositoryIssues)) / rules.GetMaxCountThreads()
	if lengthOfPartComparableIssuesSlice <= 1 {
		comparator.iterating(
			repositoryIssues,
			comparableIssues,
			int64(0),
			int64(len(repositoryIssues)),
			rules,
			result,
			intersections,
			nil,
		)
		runtimeinfo.LogInfo("RUN TASK [", taskKey, "] IN ONE THREAD.")
	} else {
		to = lengthOfPartComparableIssuesSlice
		runtimeinfo.LogInfo("RUN TASK [", taskKey, "] IN MULT. THREAD.")
		for {
			runtimeinfo.LogInfo("\t\t\t->RUN TASK [", taskKey, "] IN NEXT THREAD.")
			if to >= int64(len(repositoryIssues)) {
				wg.Add(1)
				go comparator.iterating(
					repositoryIssues,
					comparableIssues,
					from,
					int64(len(repositoryIssues)),
					rules,
					result,
					intersections,
					wg,
				)
				break
			}
			wg.Add(1)
			go comparator.iterating(
				repositoryIssues,
				comparableIssues,
				from,
				to,
				rules,
				result,
				intersections,
				wg,
			)
			from = to + 1
			to = from + lengthOfPartComparableIssuesSlice
		}
		wg.Wait()
	}
	for comparableRepositoryID, intersection := range intersections {
		allCountIntersections := int64(0)
		for _, issueIntersections := range intersection {
			allCountIntersections += issueIntersections
		}
		maxCountIntersections := countIssuesRepository * int64(len(intersection))
		coefficient := float64(allCountIntersections) / float64(maxCountIntersections)
		intersectionModels = append(
			intersectionModels,
			dataModel.NumberIssueIntersectionsModel{
				Model:                  gorm.Model{},
				RepositoryID:           repositoryID,
				ComparableRepositoryID: comparableRepositoryID,
				NumberIntersections:    coefficient,
				//
				RepositoryCountIssues: countIssuesRepository,
				CountNearestPairs:     allCountIntersections,
			},
		)
	}
	err = comparator.db.AddNumbersIntersections(intersectionModels)
	if err != nil {
		result.err = err
	}
	runtimeinfo.LogInfo("FINISH COMPARE TASK [", taskKey, "].")
	rules.GetReturnResult()(result)
	return
}
