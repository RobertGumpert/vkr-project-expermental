package issuesComparator

import (
	"issue-indexer/app/models/dataModel"
	"issue-indexer/pckg/runtimeinfo"
	"sync"
)

func (comparator *IssuesComparator) runCompareIssuesInPairs(main, second []dataModel.Issue, issueComparator IssuesInPairComparator) chan bool {
	var (
		channelFinishCompare = make(chan bool)
	)
	go func() {
		runtimeinfo.LogInfo("COMPARING START FOR : main [", main[0].RepositoryID, "], second [", second[0].RepositoryID, "]")
		wg := new(sync.WaitGroup)
		sliceLen := len(main) / comparator.MaxCountThreads
		if sliceLen <= 1 {
			comparator.compareIssuesInPairs(main, second, issueComparator, nil)
		} else {
			to := sliceLen - 1
			from := 0
			for {
				if (from + to) >= len(main) {
					wg.Add(1)
					piece := main[from : len(main)-1]
					go comparator.compareIssuesInPairs(piece, second, issueComparator, wg)
					break
				}
				wg.Add(1)
				piece := main[from:to]
				go comparator.compareIssuesInPairs(piece, second, issueComparator, wg)
				from = from + to + 1
				to = from + (sliceLen - 1)
			}
			wg.Wait()
		}
		runtimeinfo.LogInfo("COMPARING FINISHED FOR : main [", main[0].RepositoryID, "], second [", second[0].RepositoryID, "]")
		channelFinishCompare <- true
		runtimeinfo.LogInfo("SEND SIGNAL FINISH FOR : main [", main[0].RepositoryID, "], second [", second[0].RepositoryID, "]")

	}()
	return channelFinishCompare
}
