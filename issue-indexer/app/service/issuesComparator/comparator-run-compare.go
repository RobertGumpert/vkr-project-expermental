package issuesComparator

import (
	"issue-indexer/app/models/dataModel"
	"issue-indexer/pckg/runtimeinfo"
	"issue-indexer/pckg/task"
	"sync"
)

func (comparator *IssuesComparator) runCompareIssuesInPairs(comparable, comparableWith []dataModel.Issue, iTask task.ITask, issueComparator ComparatorIssuesInPair) {
	runtimeinfo.LogInfo("START COMPARE ISSUES IN PAIRS FOR : comparable [", comparable[0].RepositoryID, "], compareWith [", comparableWith[0].RepositoryID, "]")
	wg := new(sync.WaitGroup)
	sliceLen := len(comparable) / comparator.MaxCountThreads
	if sliceLen <= 1 {
		comparator.iterateIssuesInPairs(comparable, comparableWith, iTask, issueComparator, nil)
	} else {
		to := sliceLen - 1
		from := 0
		for {
			if (from + to) >= len(comparable) {
				wg.Add(1)
				piece := comparable[from : len(comparable)-1]
				go comparator.iterateIssuesInPairs(piece, comparableWith, iTask, issueComparator, wg)
				break
			}
			wg.Add(1)
			piece := comparable[from:to]
			go comparator.iterateIssuesInPairs(piece, comparableWith, iTask, issueComparator, wg)
			from = from + to + 1
			to = from + (sliceLen - 1)
		}
		wg.Wait()
	}
	iTask.SetExecutionStatus(true)
	runtimeinfo.LogInfo("FINISH COMPARE ISSUES IN PAIRS FOR : : comparable [", comparable[0].RepositoryID, "], compareWith [", comparableWith[0].RepositoryID, "]")
	comparator.channelSendResultCompare <- iTask
	runtimeinfo.LogInfo("SEND SIGNAL FOR COMPARE ISSUES IN PAIRS FOR : : comparable [", comparable[0].RepositoryID, "], compareWith [", comparableWith[0].RepositoryID, "]")
}
