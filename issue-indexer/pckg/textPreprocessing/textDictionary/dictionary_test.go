package textDictionary

import (
	"issue-indexer/pckg/runtimeinfo"
	"issue-indexer/pckg/textPreprocessing"
	"testing"
)

var(
	corpus = []string{
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

func TestFlowFullDictionary(t *testing.T) {
	dictionary, vectorsOfWords, count := FullDictionary(corpus, textPreprocessing.ParallelMode)
	runtimeinfo.LogInfo(count)
	runtimeinfo.LogInfo(vectorsOfWords)
	for item := range dictionary.IterBuffered() {
		runtimeinfo.LogInfo("[",item.Key, "] = [", item.Val, "]")
	}
	//
	dictionary, vectorsOfWords, count = FullDictionary(corpus, textPreprocessing.LinearMode)
	runtimeinfo.LogInfo(count)
	runtimeinfo.LogInfo(vectorsOfWords)
	for item := range dictionary.IterBuffered() {
		runtimeinfo.LogInfo("[",item.Key, "] = [", item.Val, "]")
	}
}

func TestFlowIDFDictionary(t *testing.T) {
	dictionary, vectorsOfWords, count := IDFDictionary(corpus, 2, textPreprocessing.ParallelMode)
	runtimeinfo.LogInfo(count)
	runtimeinfo.LogInfo(vectorsOfWords)
	for item := range dictionary.IterBuffered() {
		runtimeinfo.LogInfo("[",item.Key, "] = [", item.Val, "]")
	}
	//
	dictionary, vectorsOfWords, count = IDFDictionary(corpus, 2, textPreprocessing.LinearMode)
	runtimeinfo.LogInfo(count)
	runtimeinfo.LogInfo(vectorsOfWords)
	for item := range dictionary.IterBuffered() {
		runtimeinfo.LogInfo("[",item.Key, "] = [", item.Val, "]")
	}
}
