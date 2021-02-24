package textMetrics

import (
	cmap "github.com/streamrail/concurrent-map"
	"sync"
)

func CreateDictionary(documents *[]*[]string) *cmap.ConcurrentMap {
	var (
		wg         = new(sync.WaitGroup)
		buffer     = cmap.New()
		dictionary = cmap.New()
	)
	for document := 0; document < len(*documents); document++ {
		if (*documents)[document] == nil {
			continue
		}
		wg.Add(1)
		go func(buffer *cmap.ConcurrentMap, document *[]string, wg *sync.WaitGroup) {
			defer wg.Done()
			for word := 0; word < len(*document); word++ {
				w := (*document)[word]
				if exist := buffer.Has(w); !exist {
					buffer.Set(w, struct{}{})
				}
			}
			return
		}(&buffer, (*documents)[document], wg)
	}
	wg.Wait()
	keys := buffer.Keys()
	for wordIndex := 0; wordIndex < len(keys); wordIndex++ {
		dictionary.Set(keys[wordIndex], int64(wordIndex))
	}
	buffer = nil
	return &dictionary
}
