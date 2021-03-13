package textVectorized

import (
	"issue-indexer/pckg/runtimeinfo"
	"issue-indexer/pckg/textPreprocessing/textDictionary"
	"testing"
)

func TestVectorizedPairDictionariesFlow(t *testing.T) {
	runtimeinfo.LogInfo("FEATURES: ")
	textA := textDictionary.TextTransformToFeaturesSlice(testCorpus[0])
	textB := textDictionary.TextTransformToFeaturesSlice(testCorpus[1])
	runtimeinfo.LogInfo(textA)
	runtimeinfo.LogInfo(textB)
	//
	runtimeinfo.LogInfo("Frequency Map: ")
	dictA := GetFrequencyMap(textA)
	dictB := GetFrequencyMap(textB)
	jsonA, _ := dictA.MarshalJSON()
	jsonB, _ := dictB.MarshalJSON()
	runtimeinfo.LogInfo(string(jsonA))
	runtimeinfo.LogInfo(string(jsonB))
	//
	bagOfWords, dictionary, intersections := VectorizedPairDictionaries(dictA, dictB)
	//
	runtimeinfo.LogInfo("DICTIONARY: ")
	jsonDICT, _ := dictionary.MarshalJSON()
	runtimeinfo.LogInfo(string(jsonDICT))
	//
	runtimeinfo.LogInfo("BAG OF WORDS: ")
	for _, bag := range bagOfWords {
		runtimeinfo.LogInfo(bag)
	}
	//
	runtimeinfo.LogInfo("INTERSECTIONS: ")
	for _, word := range intersections {
		runtimeinfo.LogInfo(word)
	}
}
