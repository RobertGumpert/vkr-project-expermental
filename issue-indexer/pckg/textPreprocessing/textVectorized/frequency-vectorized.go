package textVectorized

import (
	concurrentMap "github.com/streamrail/concurrent-map"
	"issue-indexer/pckg/textPreprocessing"
	"sync"
)


func FrequencyVectorized(corpus [][]string, dictionary concurrentMap.ConcurrentMap, mode textPreprocessing.ThreadMode) [][]float64 {
	if mode == textPreprocessing.ParallelMode {
		return parallelCreateFrequencyVectors(corpus, dictionary)
	}
	return linearCreateFrequencyVectors(corpus, dictionary)
}

func linearCreateFrequencyVectors(corpus [][]string, dictionary concurrentMap.ConcurrentMap) [][]float64 {
	var (
		bagOfWords = make([][]float64, len(corpus))
	)
	for text := 0; text < len(corpus); text++ {
		transformToFrequencyVector(corpus[text], text, dictionary, bagOfWords)
	}
	return bagOfWords
}

func parallelCreateFrequencyVectors(corpus [][]string, dictionary concurrentMap.ConcurrentMap) [][]float64 {
	var (
		wg         = new(sync.WaitGroup)
		bagOfWords = make([][]float64, len(corpus))
	)
	for text := 0; text < len(corpus); text++ {
		wg.Add(1)
		func(text []string, positionVectorInCorpus int, dictionary concurrentMap.ConcurrentMap, bagOfWords [][]float64, wg*sync.WaitGroup) {
			defer wg.Done()
			transformToFrequencyVector(text, positionVectorInCorpus, dictionary, bagOfWords)
			return
		}(corpus[text], text, dictionary, bagOfWords, wg)
	}
	wg.Wait()
	return bagOfWords
}

func getFrequenciesWordsInText(text []string) concurrentMap.ConcurrentMap {
	vector := concurrentMap.New()
	for word := 0; word < len(text); word++ {
		if item, exist := vector.Get(text[word]); !exist {
			vector.Set(text[word], float64(1))
		} else {
			freq := item.(float64) + 1
			vector.Upsert(text[word], freq, func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
				return newValue
			})
		}
	}
	return vector
}

func transformToFrequencyVector(text []string, positionVectorInBagOfWords int, dictionary concurrentMap.ConcurrentMap, bagOfWords [][]float64) {
	frequencyVector := make([]float64, dictionary.Count())
	wordsVector := getFrequenciesWordsInText(text)
	for dictionaryItem := range dictionary.IterBuffered() {
		positionInFrequencyVector := dictionaryItem.Val.(int64)
		if item, exist := wordsVector.Get(dictionaryItem.Key); !exist {
			frequencyVector[positionInFrequencyVector] = float64(0)
		} else {
			frequencyVector[positionInFrequencyVector] = item.(float64)
		}
	}
	bagOfWords[positionVectorInBagOfWords] = frequencyVector
}