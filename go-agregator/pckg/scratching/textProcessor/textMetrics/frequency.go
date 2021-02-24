package textMetrics

import (
	cmap "github.com/streamrail/concurrent-map"
	"strings"
	"sync"
)

func WordsFrequencyTF(document *[]string) *cmap.ConcurrentMap {
	wg := new(sync.WaitGroup)
	wordsFrequency := cmap.New()
	length := len(*document)
	for word := 0; word < length; word++ {
		var w = (*document)[word]
		if wordsFrequency.Has(w) {
			continue
		} else {
			wordsFrequency.Set(w, 0)
		}
		wg.Add(1)
		go func(document *[]string, length float64, word int, wg *sync.WaitGroup) {
			defer wg.Done()
			frequency := float64(0)
			for otherWord := 0; otherWord < len(*document); otherWord++ {
				if strings.Compare((*document)[word], (*document)[otherWord]) == 0 {
					frequency++
				}
			}
			wordsFrequency.Set((*document)[word], frequency/length)
			return
		}(document, float64(length), word, wg)
	}
	wg.Wait()
	return &wordsFrequency
}

func WordsFrequency(document *[]string) *cmap.ConcurrentMap {
	wg := new(sync.WaitGroup)
	wordsFrequency := cmap.New()
	length := len(*document)
	for word := 0; word < length; word++ {
		var w = (*document)[word]
		if wordsFrequency.Has(w) {
			continue
		}
		wg.Add(1)
		go func(document *[]string, word int, wg *sync.WaitGroup) {
			defer wg.Done()
			frequency := float64(0)
			for otherWord := 0; otherWord < len(*document); otherWord++ {
				if strings.Compare((*document)[word], (*document)[otherWord]) == 0 {
					frequency++
				}
			}
			wordsFrequency.Set((*document)[word], frequency)
			return
		}(document, word, wg)
	}
	wg.Wait()
	return &wordsFrequency
}

