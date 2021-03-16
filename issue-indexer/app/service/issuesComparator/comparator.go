package issuesComparator

import (
	"issue-indexer/app/models/dataModel"
	"issue-indexer/pckg/task"
)


type ComparatorIssuesInPair func(i, j int, comparable, comparableWith []dataModel.Issue) (dataModel.NearestIssues, error)
type GettingResult func(iTask task.ITask)

type IssuesComparator struct {
	MaxCountThreads                  int
	MinimumTextCompletenessThreshold float64
	//
	gettingResult            GettingResult
	channelSendResultCompare chan task.ITask
}

func NewComparator(maxCountThreads int, minimumCompletenessThreshold float64, gettingResult GettingResult) *IssuesComparator {
	indexer := &IssuesComparator{
		MaxCountThreads:                  maxCountThreads,
		MinimumTextCompletenessThreshold: minimumCompletenessThreshold,
		//
		gettingResult:            gettingResult,
		channelSendResultCompare: make(chan task.ITask),
	}
	go indexer.scanResultCompareChannel()
	return indexer
}

func (comparator *IssuesComparator) AddCompareIssuesInPairs(comparable, comparableWith []dataModel.Issue, iTask task.ITask, issueComparator ComparatorIssuesInPair) {
	go comparator.runCompareIssuesInPairs(
		comparable,
		comparableWith,
		iTask,
		issueComparator,
	)
	return
}

func (comparator *IssuesComparator) scanResultCompareChannel() {
	for taskState := range comparator.channelSendResultCompare {
		comparator.gettingResult(taskState)
	}
}
