package issuesComparator

import (
	"issue-indexer/app/models/dataModel"
	"issue-indexer/pckg/runtimeinfo"
)

type IssuesInPairComparator func(i, j int, main, second []dataModel.Issue) (dataModel.NearestIssues, error)
type GettingResult func(compareResult interface{})

type IssuesComparator struct {
	MaxChannelBufferSize             int
	MaxCountThreads                  int
	MinimumTextCompletenessThreshold float64
	//
	gettingResult            GettingResult
	channelSendCompareResult chan interface{}
}

func NewComparator(maxChannelBufferSize, maxCountThreads int, minimumCompletenessThreshold float64, gettingResult GettingResult) *IssuesComparator {
	indexer := &IssuesComparator{
		MaxChannelBufferSize:             maxChannelBufferSize,
		MaxCountThreads:                  maxCountThreads,
		MinimumTextCompletenessThreshold: minimumCompletenessThreshold,
		//
		gettingResult:            gettingResult,
		channelSendCompareResult: make(chan interface{}, maxChannelBufferSize),
	}
	go indexer.scanResultCompareChannel()
	return indexer
}

func (comparator *IssuesComparator) AddCompareIssuesInPairs(main, second []dataModel.Issue, issueComparator IssuesInPairComparator) chan bool {
	return comparator.runCompareIssuesInPairs(
		main,
		second,
		issueComparator,
	)
}

func (comparator *IssuesComparator) scanResultCompareChannel() {
	for compareResult := range comparator.channelSendCompareResult {
		runtimeinfo.LogInfo("GETTING RESULT...")
		comparator.gettingResult(compareResult)
	}
}
