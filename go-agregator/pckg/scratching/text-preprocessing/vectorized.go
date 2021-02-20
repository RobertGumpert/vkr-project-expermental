package text_preprocessing

import (
	"errors"
	"fmt"
	concurrentMap "github.com/streamrail/concurrent-map"
	"go-agregator/pckg/runtimeinfo"
	"sync"
)

type VectorizedCorpusModel struct {
	Key            string
	FrequencyWords *concurrentMap.ConcurrentMap
	//
	// DOC: to dictionaryCorpusSlice
	//
	// DictionaryCorpus -
	// массив всех слов в
	// корпусе текстов.
	//
	// 		   * text-1 -> "c c c d d d d d e e e g g g g g g g"
	// 		   * text-2 -> "a a b c g g g"
	//
	// dictionaryCorpusSlice -> [a, b, c, d, e, g]
	//
	dictionaryCorpusSlice *[]*string
	dictionaryCorpusMap   *concurrentMap.ConcurrentMap
	//
	// DOC: to wordsFrequencyVectors.
	// Значениями массива float является частота
	// слова в тексте (НЕ В КОРУСЕ ТЕКСТОВ).
	//
	// dictionaryCorpusSlice -> [a, b, c, d, e, g]
	//						|  |  |  |  |  |
	// "text-1-key"		-> [0, 0, 3, 5, 3, 7]
	//						|  |  |  |  |  |
	// "text-2-key"		-> [2, 1, 1, 0, 0, 3]
	//
	wordsFrequencyVectors *concurrentMap.ConcurrentMap
	//
	// [
	//	  [0, 0, 3, 5, 3, 7],
	//    [2, 1, 1, 0, 0, 3]
	// ]
	//
	wordsFrequencyMatrix *[]*[]float64
	//
	// DOC: to wordsPresenceVectors.
	// Значениями массива float является флаг,
	// означающий, что слово присутсвует
	// в тексте 1 or 0 (НЕ В КОРУСЕ ТЕКСТОВ).
	//
	// dictionaryCorpusSlice -> [a, b, c, d, e, g]
	//							 |  |  |  |  |  |
	// "text-1-key"			 -> [0, 0, 1, 1, 1, 1]
	//							 |  |  |  |  |  |
	// "text-2-key"			 -> [1, 1, 1, 0, 0, 1]
	//
	wordsPresenceVectors *concurrentMap.ConcurrentMap
	//
	// [
	//	  [0, 0, 1, 1, 1, 1],
	//    [1, 1, 1, 0, 0, 1]
	// ]
	//
	wordsPresenceMatrix *[]*[]float64
}

func (vcm *VectorizedCorpusModel) GetFrequencyVector(key string) (*[]float64, error) {
	if vector, exist := vcm.wordsFrequencyVectors.Get(key); exist {
		return vector.(*[]float64), nil
	}
	return nil, errors.New("Vector isn't exist. ")
}

func (vcm *VectorizedCorpusModel) GetFrequencyVectors() *concurrentMap.ConcurrentMap {
	return vcm.wordsPresenceVectors
}

func (vcm *VectorizedCorpusModel) GetPresenceVectors() *concurrentMap.ConcurrentMap {
	return vcm.wordsFrequencyVectors
}

func (vcm *VectorizedCorpusModel) GetPresenceVector(key string) (*[]float64, error) {
	if vector, exist := vcm.wordsPresenceVectors.Get(key); exist {
		return vector.(*[]float64), nil
	}
	return nil, errors.New("Vector isn't exist. ")
}

func (vcm *VectorizedCorpusModel) GetMatrices() (*[]*[]float64, *[]*[]float64) {
	return vcm.wordsPresenceMatrix, vcm.wordsFrequencyMatrix
}

func (vcm *VectorizedCorpusModel) GetCorpusDictionary() (*[]*string, error) {
	if len(*vcm.dictionaryCorpusSlice) != 0 {
		return vcm.dictionaryCorpusSlice, nil
	}
	return nil, errors.New("Dictionary is empty. ")
}

func CreateDictionaryFromCorpus(corpus ...*VectorizedCorpusModel) (*[]*string, *concurrentMap.ConcurrentMap) {
	var (
		wg                    = new(sync.WaitGroup)
		mx                    = new(sync.Mutex)
		buffer                = concurrentMap.New()
		dictionaryCorpusMap   = concurrentMap.New()
		dictionaryCorpusSlice = make([]*string, 0)
	)
	for i := 0; i < len(corpus); i++ {
		wg.Add(1)
		go func(textDictionary, buffer *concurrentMap.ConcurrentMap,
			dictionaryCorpusSlice *[]*string,
			wg *sync.WaitGroup, mx *sync.Mutex) {
			defer wg.Done()
			words := textDictionary.Keys()
			for j := 0; j < len(words); j++ {
				word := words[j]
				if exist := buffer.Has(word); !exist {
					buffer.Set(word, struct{}{})
					mx.Lock()
					*dictionaryCorpusSlice = append(*dictionaryCorpusSlice, &word)
					mx.Unlock()
				}
			}
			return
		}(
			corpus[i].FrequencyWords,
			&buffer,
			&dictionaryCorpusSlice,
			wg,
			mx,
		)
	}
	wg.Wait()
	for i := 0; i < len(dictionaryCorpusSlice); i++ {
		dictionaryCorpusMap.Set(*dictionaryCorpusSlice[i], int64(i))
	}
	return &dictionaryCorpusSlice, &dictionaryCorpusMap
}

func VectorizedWithDictionary(dictionaryCorpusMap *concurrentMap.ConcurrentMap, corpus ...*VectorizedCorpusModel) *VectorizedCorpusModel {
	model := vectorizedCorpus(nil, dictionaryCorpusMap, corpus...)
	return model
}

func VectorizedCorpus(corpus ...*VectorizedCorpusModel) *VectorizedCorpusModel {
	dictionaryCorpusSlice, dictionaryCorpusMap := CreateDictionaryFromCorpus(corpus...)
	model := vectorizedCorpus(dictionaryCorpusSlice, dictionaryCorpusMap, corpus...)
	return model
}

func vectorizedCorpus(dictionaryCorpusSlice *[]*string, dictionaryCorpusMap *concurrentMap.ConcurrentMap, corpus ...*VectorizedCorpusModel) *VectorizedCorpusModel {
	var (
		wg    = new(sync.WaitGroup)
		mx    = new(sync.Mutex)
		model = &VectorizedCorpusModel{
			Key:            "vectorized-corpus-result",
			FrequencyWords: nil,
		}
		wordsFrequencyVectors = concurrentMap.New()
		wordsPresenceVectors  = concurrentMap.New()
		wordsFrequencyMatrix  = make([]*[]float64, len(corpus))
		wordsPresenceMatrix   = make([]*[]float64, len(corpus))
	)
	if dictionaryCorpusSlice == nil {
		wg.Add(1)
		go func(dictionaryCorpusSlice *[]*string, dictionaryCorpusMap *concurrentMap.ConcurrentMap, wg *sync.WaitGroup) {
			defer wg.Done()
			slice := make([]*string, dictionaryCorpusMap.Count())
			for item := range dictionaryCorpusMap.IterBuffered() {
				word := item.Key
				index := item.Val.(int64)
				slice[index] = &word
			}
			dictionaryCorpusSlice = &slice
			return
		}(dictionaryCorpusSlice, dictionaryCorpusMap, wg)
	}
	for i := 0; i < len(corpus); i++ {
		wg.Add(1)
		go func(i int, text *VectorizedCorpusModel,
			wordsFrequencyMatrix, wordsPresenceMatrix *[]*[]float64,
			wordsFrequencyVectors, wordsPresenceVectors, dictionaryCorpusMap *concurrentMap.ConcurrentMap,
			wg *sync.WaitGroup, mx *sync.Mutex) {
			defer wg.Done()
			var (
				l               = dictionaryCorpusMap.Count()
				vectorFrequency = make([]float64, l)
				vectorPresence  = make([]float64, l)
			)
			mx.Lock()
			(*wordsFrequencyMatrix)[i] = &vectorFrequency
			(*wordsPresenceMatrix)[i] = &vectorPresence
			mx.Unlock()
			for item := range dictionaryCorpusMap.IterBuffered() {
				var (
					key = item.Key
					j   = item.Val.(int64)
				)
				if text.FrequencyWords.Has(key) {
					value, _ := text.FrequencyWords.Get(key)
					wordFrequency := switchToFloat64(value)
					mx.Lock()
					(*(*wordsFrequencyMatrix)[i])[j] = wordFrequency
					(*(*wordsPresenceMatrix)[i])[j] = 1
					mx.Unlock()
				} else {
					mx.Lock()
					(*(*wordsFrequencyMatrix)[i])[j] = 0
					(*(*wordsPresenceMatrix)[i])[j] = 0
					mx.Unlock()
				}
			}
			wordsPresenceVectors.Set(text.Key, (*wordsPresenceMatrix)[i])
			wordsFrequencyVectors.Set(text.Key, (*wordsFrequencyMatrix)[i])
			return
		}(
			i,
			corpus[i],
			&wordsFrequencyMatrix,
			&wordsPresenceMatrix,
			&wordsFrequencyVectors,
			&wordsPresenceVectors,
			dictionaryCorpusMap,
			wg,
			mx,
		)
	}
	wg.Wait()
	model.wordsFrequencyMatrix = &wordsFrequencyMatrix
	model.wordsFrequencyVectors = &wordsFrequencyVectors
	model.wordsPresenceMatrix = &wordsPresenceMatrix
	model.wordsPresenceVectors = &wordsPresenceVectors
	model.dictionaryCorpusSlice = dictionaryCorpusSlice
	model.dictionaryCorpusMap = dictionaryCorpusMap
	return model
}

func Vectorized(vecA, vecB *concurrentMap.ConcurrentMap) (A, B *[]float64) {
	var (
		buffer = concurrentMap.New()
		vecAC  = make([]float64, 0)
		vecBC  = make([]float64, 0)
	)
	for item := range vecA.IterBuffered() {
		switchedValue := switchToFloat64(item.Val)
		vecAC = append(vecAC, switchedValue)
		if val, exist := vecB.Get(item.Key); !exist {
			vecBC = append(vecBC, float64(0))
		} else {
			switchedValue := switchToFloat64(val)
			vecBC = append(vecBC, switchedValue)
		}
		buffer.Set(item.Key, struct{}{})
	}
	for item := range vecB.IterBuffered() {
		if buffer.Has(item.Key) == false {
			switchedValue := switchToFloat64(item.Val)
			vecBC = append(vecBC, switchedValue)
			vecAC = append(vecAC, float64(0))
		}
	}
	if len(vecAC) != len(vecBC) {
		panic("Vector lengths isn't equal.")
	}
	return &vecAC, &vecBC
}

func addPanic() {
	panic(
		fmt.Sprintf(
			"%s, PANIC : %s",
			runtimeinfo.Runtime(2),
			"Type not number or not 64 size.",
		),
	)
}

func switchToFloat64(value interface{}) float64 {
	switch value.(type) {
	case int64:
		v := value.(int64)
		return float64(v)
	case float64:
		return value.(float64)
	default:
		addPanic()
	}
	return 0
}
