package textDictionary

import (
	"issue-indexer/pckg/runtimeinfo"
	"issue-indexer/pckg/textPreprocessing"
	"testing"
)

func TestFlowIDFDictionary(t *testing.T) {
	dictionary, vectorsOfWords, count := IDFDictionary(testCorpus, 2, textPreprocessing.ParallelMode)
	runtimeinfo.LogInfo(count)
	runtimeinfo.LogInfo(vectorsOfWords)
	for item := range dictionary.IterBuffered() {
		runtimeinfo.LogInfo("[",item.Key, "] = [", item.Val, "]")
	}
	//
	dictionary, vectorsOfWords, count = IDFDictionary(testCorpus, 2, textPreprocessing.LinearMode)
	runtimeinfo.LogInfo(count)
	runtimeinfo.LogInfo(vectorsOfWords)
	for item := range dictionary.IterBuffered() {
		runtimeinfo.LogInfo("[",item.Key, "] = [", item.Val, "]")
	}
}
