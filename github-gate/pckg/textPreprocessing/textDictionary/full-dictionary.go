package textDictionary

import (
	"github-gate/pckg/textPreprocessing"
	"github.com/bbalet/stopwords"
	concurrentMap "github.com/streamrail/concurrent-map"
	"strings"
	"sync"
)

func TextTransformToFeaturesSlice(text string) []string {
	text = strings.ToLower(text)
	if strings.Contains(text, "\n") {
		text = strings.ReplaceAll(text, "\n", " ")
	}
	if strings.Contains(text, "\r") {
		text = strings.ReplaceAll(text, "\r", " ")
	}
	text = stopwords.CleanString(text, "en", true)
	allWords := strings.Split(text, " ")
	words := make([]string, 0)
	for i, word := range allWords {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}
		words = append(
			words,
			allWords[i],
		)
	}
	return words
}

//
//
//
//------------------FULL DICTIONARY-------------------------------------------------------------------------------------
//
//
//

func FullDictionary(corpus []string, mode textPreprocessing.ThreadMode) (concurrentMap.ConcurrentMap, [][]string, int) {
	if mode == textPreprocessing.ParallelMode {
		return parallelCreateFullDictionary(corpus)
	}
	return linearCreateFullDictionary(corpus)
}

func linearCreateFullDictionary(corpus []string) (concurrentMap.ConcurrentMap, [][]string, int) {
	var (
		dictionary     = concurrentMap.New()
		vectorsOfWords = make([][]string, len(corpus))
	)
	for text := 0; text < len(corpus); text++ {
		if corpus[text] == "" || corpus[text] == "\n" || corpus[text] == "\r" || corpus[text] == "\r\n" {
			continue
		}
		addWordsToFullDictionary(dictionary, corpus[text], text, vectorsOfWords)
	}
	dictionaryTransform(dictionary)
	return dictionary, vectorsOfWords, dictionary.Count()
}

func parallelCreateFullDictionary(corpus []string) (concurrentMap.ConcurrentMap, [][]string, int) {
	var (
		wg             = new(sync.WaitGroup)
		dictionary     = concurrentMap.New()
		vectorsOfWords = make([][]string, len(corpus))
	)
	for text := 0; text < len(corpus); text++ {
		if corpus[text] == "" || corpus[text] == "\n" || corpus[text] == "\r" || corpus[text] == "\r\n" {
			continue
		}
		wg.Add(1)
		go func(dictionary concurrentMap.ConcurrentMap, text string, positionTextInCorpus int, vectorsOfWords [][]string, wg *sync.WaitGroup) {
			defer wg.Done()
			addWordsToFullDictionary(dictionary, text, positionTextInCorpus, vectorsOfWords)
			return
		}(dictionary, corpus[text], text, vectorsOfWords, wg)
	}
	wg.Wait()
	dictionaryTransform(dictionary)
	return dictionary, vectorsOfWords, dictionary.Count()
}

func addWordsToFullDictionary(dictionary concurrentMap.ConcurrentMap, text string, positionInVectorsOfWords int, vectorsOfWords [][]string) {
	words := TextTransformToFeaturesSlice(text)
	if len(words) == 0 {
		return
	}
	for word := 0; word < len(words); word++ {
		vectorsOfWords[positionInVectorsOfWords] = append(
			vectorsOfWords[positionInVectorsOfWords],
			words[word],
		)
		if exist := dictionary.Has(words[word]); exist {
			continue
		} else {
			dictionary.Set(words[word], struct{}{})
		}
	}
}

func dictionaryTransform(dictionary concurrentMap.ConcurrentMap) {
	words := dictionary.Keys()
	for position := 0; position < len(words); position++ {
		dictionary.Upsert(words[position], int64(position), func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
			return newValue
		})
	}
}
