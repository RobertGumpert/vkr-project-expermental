package issuesComparator

import (
	"issue-indexer/app/models/dataModel"
	"issue-indexer/pckg/task"
	"runtime"
	"sync"
)

func (comparator *IssuesComparator) iterateIssuesInPairs(comparable, comparableWith []dataModel.Issue, iTask task.ITask, issueComparator ComparatorIssuesInPair, wg *sync.WaitGroup) {
	for i := 0; i < len(comparable); i++ {
		for j := 0; j < len(comparableWith); j++ {
			nearestIssues, err := issueComparator(i, j, comparable, comparableWith)
			if err != nil {
				continue
			}
			iTask.SetResult(nearestIssues)
			comparator.channelSendResultCompare <- iTask
		}
	}
	runtime.GC()
	if wg != nil {
		wg.Done()
	}
	return
}
