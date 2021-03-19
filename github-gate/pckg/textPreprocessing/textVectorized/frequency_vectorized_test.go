package textVectorized

import (
	"issue-indexer/pckg/runtimeinfo"
	"issue-indexer/pckg/textPreprocessing"
	"issue-indexer/pckg/textPreprocessing/textDictionary"
	"testing"
)

var(
	testCorpus = []string{
		// Vue
		"Vue js is a progressive incrementally adoptable JavaScript framework for building UI on the web\r\nframework frontend javascript vue\n",
		// React
		"A declarative efficient and flexible JavaScript library for building user interfaces declarative frontend javascript library react ui",
		//Hyper
		"A terminal built on web technologies css html hyper javascript linux macos react terminal terminal emulators",
		// Alacritty
		"A cross platform OpenGL terminal emulator bsd gpu linux macos opengl rust terminal terminal emulators vte windows",
	}
)

func TestFullDictionaryFrequencyVectorizedFlow(t *testing.T) {
	dictionary, vectorsOfWords, countFeatures := textDictionary.FullDictionary(testCorpus, textPreprocessing.LinearMode)
	//
	runtimeinfo.LogInfo(countFeatures)
	bagOfWords := FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.LinearMode)
	for _, bag := range bagOfWords {
		runtimeinfo.LogInfo(bag)
	}
	//
	runtimeinfo.LogInfo(countFeatures)
	bagOfWords = FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.ParallelMode)
	for _, bag := range bagOfWords {
		runtimeinfo.LogInfo(bag)
	}
}

func TestIDFDictionaryFrequencyVectorizedFlow(t *testing.T) {
	dictionary, vectorsOfWords, countFeatures := textDictionary.IDFDictionary(testCorpus, 2, textPreprocessing.LinearMode)
	//
	runtimeinfo.LogInfo(countFeatures)
	bagOfWords := FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.LinearMode)
	for _, bag := range bagOfWords {
		runtimeinfo.LogInfo(bag)
	}
	//
	runtimeinfo.LogInfo(countFeatures)
	bagOfWords = FrequencyVectorized(vectorsOfWords, dictionary, textPreprocessing.ParallelMode)
	for _, bag := range bagOfWords {
		runtimeinfo.LogInfo(bag)
	}
}