package textMetrics

import (
	"issue-indexer/pckg/runtimeinfo"
	"issue-indexer/pckg/textPreprocessing"
	"issue-indexer/pckg/textPreprocessing/textDictionary"
	"issue-indexer/pckg/textPreprocessing/textVectorized"
	"testing"
)

func TestFullDictionaryIntersectionsFlow(t *testing.T) {
	dictionary, vectorsOfWords, countFeatures := textDictionary.FullDictionary(testCorpus, textPreprocessing.LinearMode)
	bagOfWords := textVectorized.FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.LinearMode)
	//
	runtimeinfo.LogInfo(countFeatures)
	intersectionsMatrix := Intersections(bagOfWords, textPreprocessing.LinearMode)
	for _, intersections := range intersectionsMatrix {
		runtimeinfo.LogInfo(intersections)
	}
	//
	runtimeinfo.LogInfo(countFeatures)
	intersectionsMatrix = Intersections(bagOfWords, textPreprocessing.ParallelMode)
	for _, intersections := range intersectionsMatrix {
		runtimeinfo.LogInfo(intersections)
	}
}

func TestIDFDictionaryIntersectionsFlow(t *testing.T) {
	dictionary, vectorsOfWords, countFeatures := textDictionary.IDFDictionary(testCorpus, 2, textPreprocessing.LinearMode)
	bagOfWords := textVectorized.FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.LinearMode)
	//
	runtimeinfo.LogInfo(countFeatures)
	intersectionsMatrix := Intersections(bagOfWords, textPreprocessing.LinearMode)
	for _, intersections := range intersectionsMatrix {
		runtimeinfo.LogInfo(intersections)
	}
	//
	runtimeinfo.LogInfo(countFeatures)
	intersectionsMatrix = Intersections(bagOfWords, textPreprocessing.ParallelMode)
	for _, intersections := range intersectionsMatrix {
		runtimeinfo.LogInfo(intersections)
	}
}

func TestIntersectionsOnPair(t *testing.T) {
	corpus := []string{
		testCorpus[0],
		testCorpus[1],
	}
	dictionary, vectorsOfWords, countFeatures := textDictionary.FullDictionary(corpus, textPreprocessing.LinearMode)
	bagOfWords := textVectorized.FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.LinearMode)
	//
	for _, bag := range bagOfWords {
		runtimeinfo.LogInfo(bag)
	}
	runtimeinfo.LogInfo("Count features: ", countFeatures)
	intersectionsMatrix := Intersections(bagOfWords, textPreprocessing.LinearMode)
	for _, intersections := range intersectionsMatrix {
		runtimeinfo.LogInfo(intersections)
	}
}
