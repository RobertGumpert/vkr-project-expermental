package textMetrics

import (

	"issue-indexer/pckg/runtimeinfo"
	"issue-indexer/pckg/textPreprocessing"
	"issue-indexer/pckg/textPreprocessing/textDictionary"
	"issue-indexer/pckg/textPreprocessing/textVectorized"
	"testing"
)

func TestFullDictionaryCompletenessTextFlow(t *testing.T) {
	dictionary, vectorsOfWords, countFeatures := textDictionary.FullDictionary(testCorpus, textPreprocessing.LinearMode)
	bagOfWords := textVectorized.FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.LinearMode)
	//
	runtimeinfo.LogInfo("Count features : ", countFeatures)
	completenessMatrix := CompletenessText(bagOfWords, textPreprocessing.LinearMode)
	for _, completeness := range completenessMatrix {
		runtimeinfo.LogInfo(completeness)
	}
	//
	runtimeinfo.LogInfo("Count features : ", countFeatures)
	completenessMatrix = CompletenessText(bagOfWords, textPreprocessing.ParallelMode)
	for _, completeness := range completenessMatrix {
		runtimeinfo.LogInfo(completeness)
	}
}

func TestIDFDictionaryCompletenessTextFlow(t *testing.T) {
	dictionary, vectorsOfWords, countFeatures := textDictionary.IDFDictionary(testCorpus, 2, textPreprocessing.LinearMode)
	bagOfWords := textVectorized.FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.LinearMode)
	//
	runtimeinfo.LogInfo("Count features : ", countFeatures)
	completenessMatrix := CompletenessText(bagOfWords, textPreprocessing.LinearMode)
	for _, completeness := range completenessMatrix {
		runtimeinfo.LogInfo(completeness)
	}
	//
	runtimeinfo.LogInfo("Count features : ", countFeatures)
	completenessMatrix = CompletenessText(bagOfWords, textPreprocessing.ParallelMode)
	for _, completeness := range completenessMatrix {
		runtimeinfo.LogInfo(completeness)
	}
}

func TestCompletenessTextOnPair(t *testing.T) {
	corpus := []string{
		testCorpus[0],
		testCorpus[1],
	}
	//
	dictionary, vectorsOfWords, countFeatures := textDictionary.FullDictionary(corpus, textPreprocessing.LinearMode)
	bagOfWords := textVectorized.FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.LinearMode)
	for _, bag := range bagOfWords {
		runtimeinfo.LogInfo(bag)
	}
	runtimeinfo.LogInfo("Count features: ", countFeatures)
	completenessMatrix := CompletenessText(bagOfWords, textPreprocessing.LinearMode)
	for _, completeness := range completenessMatrix {
		runtimeinfo.LogInfo(completeness)
	}
	//
	distance, _ := CosineDistanceOnPairVectors(bagOfWords)
	runtimeinfo.LogInfo("Distance: ", distance * 100)
}