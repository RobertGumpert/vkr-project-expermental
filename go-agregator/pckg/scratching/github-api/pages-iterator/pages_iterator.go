package pages_iterator

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type PagesIterator struct {
	ApiRateLimit time.Duration        `json:"api_rate_limit"`
	mx           *sync.RWMutex        `json:"-"`
	Iterators    map[string]*iterator `json:"iterators"`
}

func NewPagesIterator(countRequestsPerMinute uint64, constructors ...Configurator) *PagesIterator {
	coefficient := 60 / countRequestsPerMinute
	apiRateLimit := time.Second * time.Duration(coefficient)
	if apiRateLimit <= 0 {
		apiRateLimit = 1
	}
	var (
		timeOut      = apiRateLimit*time.Duration(len(constructors)) + time.Millisecond*time.Duration(50)
		pageIterator = &PagesIterator{
			ApiRateLimit: time.Second * time.Duration(coefficient),
			mx:           new(sync.RWMutex),
			Iterators:    make(map[string]*iterator),
		}
	)
	for _, constructor := range constructors {
		iterator := constructor()
		iterator.UsedTimeOut = &timeOut
		iterator.ApiRateLimit = &apiRateLimit
		pageIterator.Iterators[iterator.IteratorKey] = iterator
	}
	return pageIterator
}

func (pagesIterator *PagesIterator) DO() {
	wg := new(sync.WaitGroup)
	for _, pg := range pagesIterator.Iterators {
		wg.Add(1)
		go pg.iterate(wg)
	}
	wg.Wait()
	return
}

func (pagesIterator *PagesIterator) Get(key string) *iterator {
	pagesIterator.mx.RLock()
	var itr = pagesIterator.Iterators[key]
	pagesIterator.mx.RUnlock()
	return itr
}

func (pagesIterator *PagesIterator) Serialize() ([]byte, error) {
	bs, err := json.Marshal(pagesIterator)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return bs, nil
}
