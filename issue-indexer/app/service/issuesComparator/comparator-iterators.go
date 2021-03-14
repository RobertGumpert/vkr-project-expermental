package issuesComparator

import (
	"issue-indexer/app/models/dataModel"
	"issue-indexer/pckg/task"
	"runtime"
	"sync"
)

func (comparator *IssuesComparator) iterateIssuesInPairs(comparable, comparableWith []dataModel.Issue, task task.ITask, issueComparator IssuesInPairComparator, wg *sync.WaitGroup) {
	//runtimeinfo.LogInfo("RUN COMPARE PIECE OF ISSUES FOR: comparable [", comparable[0].RepositoryID, "], compareWith [", comparableWith[0].RepositoryID, "]")
	for i := 0; i < len(comparable); i++ {
		for j := 0; j < len(comparableWith); j++ {
			nearestIssues, err := issueComparator(i, j, comparable, comparableWith)
			if err != nil {
				continue
			}
			task.SetResult(nearestIssues)
			comparator.channelSendResultCompare <- task
		}
	}
	runtime.GC()
	if wg != nil {
		wg.Done()
	}
	//runtimeinfo.LogInfo("FINISH COMPARE PIECE OF ISSUES FOR: comparable [", comparable[0].RepositoryID, "], compareWith [", comparableWith[0].RepositoryID, "]")
	return
}
