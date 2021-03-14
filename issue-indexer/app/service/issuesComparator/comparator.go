package issuesComparator

import (
	"issue-indexer/app/models/dataModel"
	"issue-indexer/pckg/task"
)


type IssuesInPairComparator func(i, j int, main, second []dataModel.Issue) (dataModel.NearestIssues, error)
type GettingResult func(compareResult task.ITask)

type IssuesComparator struct {
	MaxChannelBufferSize             int
	MaxCountThreads                  int
	MinimumTextCompletenessThreshold float64
	//
	gettingResult            GettingResult
	channelSendResultCompare chan task.ITask
}

func NewComparator(maxChannelBufferSize, maxCountThreads int, minimumCompletenessThreshold float64, gettingResult GettingResult) *IssuesComparator {
	indexer := &IssuesComparator{
		MaxChannelBufferSize:             maxChannelBufferSize,
		MaxCountThreads:                  maxCountThreads,
		MinimumTextCompletenessThreshold: minimumCompletenessThreshold,
		//
		gettingResult:            gettingResult,
		channelSendResultCompare: make(chan task.ITask, maxChannelBufferSize),
	}
	go indexer.scanResultCompareChannel()
	return indexer
}

func (comparator *IssuesComparator) AddCompareIssuesInPairs(comparable, comparableWith []dataModel.Issue, task task.ITask, issueComparator IssuesInPairComparator) {
	go comparator.runCompareIssuesInPairs(
		comparable,
		comparableWith,
		task,
		issueComparator,
	)
	return
}

func (comparator *IssuesComparator) scanResultCompareChannel() {
	for taskState := range comparator.channelSendResultCompare {
		comparator.gettingResult(taskState)
	}
}
