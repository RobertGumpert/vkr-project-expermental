package textVectorized

import (
	concurrentMap "github.com/streamrail/concurrent-map"
)

func VectorizedPairDictionaries(dictA, dictB concurrentMap.ConcurrentMap) ([][]float64, concurrentMap.ConcurrentMap, []string) {
	var (
		dictionary    = concurrentMap.New()
		intersections = make([]string, 0)
		index         = int64(0)
		bagOfWords    = make([][]float64, 2)
		vecA          = make([]float64, 0)
		vecB          = make([]float64, 0)
	)
	for word := range dictA.IterBuffered() {
		dictionary.Set(word.Key, index)
		vecA = append(vecA, word.Val.(float64))
		if value, exist := dictB.Get(word.Key); exist {
			vecB = append(vecB, value.(float64))
			intersections = append(
				intersections,
				word.Key,
			)
		} else {
			vecB = append(vecB, float64(0))
		}
		index++
	}
	for word := range dictB.IterBuffered() {
		if dictionary.Has(word.Key) {
			continue
		}
		dictionary.Set(word.Key, index)
		vecB = append(vecB, word.Val.(float64))
		vecA = append(vecA, float64(0))
		index++
	}
	bagOfWords[0] = vecA
	bagOfWords[1] = vecB
	return bagOfWords, dictionary, intersections
}
