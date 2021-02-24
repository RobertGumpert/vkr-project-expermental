package textVectoring

import (
	"errors"
	cmap "github.com/streamrail/concurrent-map"
	"sync"
)

func Vectorized(dictionary *cmap.ConcurrentMap, vectors *[]*cmap.ConcurrentMap) error {
	if dictionary == nil || dictionary.Count() == 0 {
		return errors.New("Dictionary is empty. ")
	}
	wg := new(sync.WaitGroup)
	for vector := 0; vector < len(*vectors); vector++ {
		wg.Add(1)
		go func(vector, dictionary *cmap.ConcurrentMap, wg *sync.WaitGroup) {
			defer wg.Done()
			for item := range dictionary.IterBuffered() {
				if !vector.Has(item.Key) {
					vector.Set(item.Key, float64(0))
				}
			}
			return
		}((*vectors)[vector], dictionary, wg)
	}
	wg.Wait()
	return nil
}
