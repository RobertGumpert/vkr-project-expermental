package issuesComparator

import (
	"issue-indexer/app/models/dataModel"
	"issue-indexer/pckg/runtimeinfo"
	"runtime"
	"sync"
)

func (comparator *IssuesComparator) compareIssuesInPairs(main, second []dataModel.Issue, issueComparator IssuesInPairComparator, wg *sync.WaitGroup) {
	runtimeinfo.LogInfo("RUN COMPARE PIECE OF ISSUES FOR: main [", main[0].RepositoryID, "], second [", second[0].RepositoryID, "]")
	for i := 0; i < len(main); i++ {
		for j := 0; j < len(second); j++ {
			nearestIssues, err := issueComparator(i, j, main, second)
			if err != nil {
				continue
			}
			comparator.channelSendCompareResult <- nearestIssues
			// runtimeinfo.LogInfo("SEND SIMILAR ISSUES: main [", main[i].RepositoryID, "], second [", second[j].RepositoryID, "]")
		}
	}
	runtime.GC()
	if wg != nil {
		wg.Done()
	}
	runtimeinfo.LogInfo("FINISH COMPARE PIECE OF ISSUES FOR: main [", main[0].RepositoryID, "], second [", second[0].RepositoryID, "]")
	return
}
