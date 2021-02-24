package textMetrics

import (
	"errors"
	cmap "github.com/streamrail/concurrent-map"
	"math"
	"sync"
)

func TFIDF(documentsTF, wordsIDF *cmap.ConcurrentMap) (*cmap.ConcurrentMap, error) {
	if documentsTF == nil || wordsIDF == nil {
		return nil, errors.New("invalid memory address or nil pointer dereference")
	}
	TFIDFWeights := cmap.New()
	for item := range documentsTF.IterBuffered() {
		if val, exist := wordsIDF.Get(item.Key); exist {
			idf := val.(float64)
			tf := item.Val.(float64)
			weight := tf * idf
			TFIDFWeights.Set(item.Key, weight)
		}
	}
	return &TFIDFWeights, nil
}

//
// TF = ni / DocLen
//	* ni - частота слова в документе.
//  * DocLen - кол-во слов в документе.
//
// IDF = log(N / dfi)
// 	* N - кол-во документов.
// 	* dfi - сколько раз слово встречается во всех документах.
//
func GetTFIDFMetrics(documents *[]*[]string) (*cmap.ConcurrentMap, *[]*cmap.ConcurrentMap, *cmap.ConcurrentMap) {
	wg := new(sync.WaitGroup)
	mx := new(sync.Mutex)
	dictionary := cmap.New()
	length := len(*documents)
	documentsTF := make([]*cmap.ConcurrentMap, length)
	wordsIDF := cmap.New()
	//
	wg.Add(1)
	go func(dictionary *cmap.ConcurrentMap, documents *[]*[]string, wg *sync.WaitGroup) {
		defer wg.Done()
		d := CreateDictionary(documents)
		*dictionary = *d
		return
	}(&dictionary, documents, wg)
	for document := 0; document < length; document++ {
		if (*documents)[document] == nil {
			continue
		}
		wg.Add(1)
		go func(document *[]string, index int, tfDocuments *[]*cmap.ConcurrentMap, wg *sync.WaitGroup, mx *sync.Mutex) {
			defer wg.Done()
			docTF := WordsFrequencyTF(document)
			mx.Lock()
			(*tfDocuments)[index] = docTF
			mx.Unlock()
			return
		}((*documents)[document], document, &documentsTF, wg, mx)
	}
	wg.Wait()
	for item := range dictionary.IterBuffered() {
		wordIDF := float64(0)
		dfi := float64(0)
		for document := 0; document < len(documentsTF); document++ {

			if documentsTF[document] == nil {
				continue
			}
			if documentsTF[document].Has(item.Key) {
				dfi++
			}
		}
		if dfi == 0 {
			dfi = float64(length)
		}
		wordIDF = math.Log(float64(length) / dfi)
		wordsIDF.Set(item.Key, wordIDF)
	}
	return &wordsIDF, &documentsTF, &dictionary
}
