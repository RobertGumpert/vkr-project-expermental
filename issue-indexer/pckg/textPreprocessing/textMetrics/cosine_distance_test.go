package textMetrics

import (
	"issue-indexer/pckg/runtimeinfo"
	"issue-indexer/pckg/textPreprocessing"
	"issue-indexer/pckg/textPreprocessing/textDictionary"
	"issue-indexer/pckg/textPreprocessing/textVectorized"
	"testing"
)

var(
	testCorpus = []string{
		// Vue
		"Vue js is a progressive incrementally adoptable JavaScript framework for building UI on the web framework frontend javascript vue",
		// React
		"A declarative efficient and flexible JavaScript library for building user interfaces declarative frontend javascript library react ui",
		//Hyper
		"A terminal built on web technologies css html hyper javascript linux macos react terminal terminal emulators",
		// Alacritty
		"A cross platform OpenGL terminal emulator bsd gpu linux macos opengl rust terminal terminal emulators vte windows",
	}
)

func TestFullDictionaryCosineDistanceFlow(t *testing.T) {
	dictionary, vectorsOfWords, countFeatures := textDictionary.FullDictionary(testCorpus, textPreprocessing.LinearMode)
	bagOfWords := textVectorized.FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.LinearMode)
	//
	runtimeinfo.LogInfo(countFeatures)
	cosineMatrix := CosineDistance(bagOfWords, textPreprocessing.LinearMode)
	for _, distance := range cosineMatrix {
		runtimeinfo.LogInfo(distance)
	}
	//
	runtimeinfo.LogInfo(countFeatures)
	cosineMatrix = CosineDistance(bagOfWords, textPreprocessing.ParallelMode)
	for _, distance := range cosineMatrix {
		runtimeinfo.LogInfo(distance)
	}
}

func TestIDFDictionaryCosineDistanceFlow(t *testing.T) {
	dictionary, vectorsOfWords, countFeatures := textDictionary.IDFDictionary(testCorpus, 2, textPreprocessing.LinearMode)
	bagOfWords := textVectorized.FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.LinearMode)
	//
	runtimeinfo.LogInfo(countFeatures)
	cosineMatrix := CosineDistance(bagOfWords, textPreprocessing.LinearMode)
	for _, distance := range cosineMatrix {
		runtimeinfo.LogInfo(distance)
	}
	//
	runtimeinfo.LogInfo(countFeatures)
	cosineMatrix = CosineDistance(bagOfWords, textPreprocessing.ParallelMode)
	for _, distance := range cosineMatrix {
		runtimeinfo.LogInfo(distance)
	}
}

func TestCosineDistanceOnPairVectorsFlow(t *testing.T) {
	corpus := []string{
		testCorpus[0],
		testCorpus[1],
	}
	//
	dictionary, vectorsOfWords, countFeatures := textDictionary.FullDictionary(corpus, textPreprocessing.LinearMode)
	bagOfWords := textVectorized.FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.LinearMode)
	runtimeinfo.LogInfo("FULL : ", countFeatures)
	if cosineDistance, err := CosineDistanceOnPairVectors(bagOfWords); err != nil {
		t.Fatal(err)
	} else {
		runtimeinfo.LogInfo(cosineDistance)
	}
	//
	dictionary, vectorsOfWords, countFeatures = textDictionary.IDFDictionary(corpus, 2, textPreprocessing.LinearMode)
	bagOfWords = textVectorized.FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.LinearMode)
	runtimeinfo.LogInfo("IDF : ", countFeatures)
	if cosineDistance, err := CosineDistanceOnPairVectors(bagOfWords); err != nil {
		t.Fatal(err)
	} else {
		runtimeinfo.LogInfo(cosineDistance)
	}
}
