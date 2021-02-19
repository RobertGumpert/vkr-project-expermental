package issues_collector

import (
	"errors"
	"fmt"
	concurrentmap "github.com/streamrail/concurrent-map"
	"go-agregator/pckg/requests"
	"go-agregator/pckg/runtimeinfo"
	mapper "go-agregator/pckg/scratching/github-api/mapper"
	pagesiterator "go-agregator/pckg/scratching/github-api/pages-iterator"
	"net/http"
	"strings"
	"sync"
)

type responsesHandler func(repositoryURL string, pagesIterator *pagesiterator.PagesIterator, wg *sync.WaitGroup, cmp *concurrentmap.ConcurrentMap)

func CustomizablePagesCollect(configurators ...Configurator) (List, error) {
	if configurators == nil || len(configurators) == 0 {
		return nil, errors.New("Configurators not created. ")
	}
	var (
		concurrentMaps            = concurrentmap.New()
		configuratorsPageIterator = make([]pagesiterator.Configurator, 0)
	)
	for _, configurator := range configurators {
		issuesIterator := configurator()
		key := strings.Split(
			issuesIterator.URL,
			"/issues?state=all",
		)[0]
		concurrentMaps.Set(
			key,
			issuesIterator,
		)
		configuratorsPageIterator = append(
			configuratorsPageIterator,
			createPagesIterator(key, issuesIterator.URL, issuesIterator.CountPages),
		)
	}
	list, err := doCollect(
		configuratorsPageIterator,
		func(repositoryURL string, pagesIterator *pagesiterator.PagesIterator, wg *sync.WaitGroup, cmp *concurrentmap.ConcurrentMap) {
			defer wg.Done()
			iterator := pagesIterator.Get(repositoryURL)
			issues := iterateHttpResponsesChannel(iterator.Responses)
			ii, _ := concurrentMaps.Get(repositoryURL)
			issuesIterator := ii.(*issuePagesIterator)
			issuesIterator.Issues = *issues
			issuesIterator.LastPage = iterator.FirstPageNumber
			cmp.Set(repositoryURL, issuesIterator)
		},
	)
	if err != nil {
		return nil, err
	}
	return List(*list), nil
}

func Collect(url string, urls ...string) (List, error) {
	if len(urls) != 0 {
		urls = append(urls, url)
	} else {
		urls = make([]string, 0)
		urls = append(urls, url)
	}
	for i := 0; i < len(urls); i++ {
		urls[i] = fmt.Sprintf("%s/issues?state=all", urls[i])
	}
	iterators := make([]pagesiterator.Configurator, 0)
	for _, repositoryURL := range urls {
		key := strings.Split(
			repositoryURL,
			"/issues?state=all",
		)[0]
		countPages, err := findCountPages(repositoryURL)
		if err != nil {
			fmt.Println(runtimeinfo.Runtime(1), "; ERROR: ", err)
			continue
		}
		iterators = append(
			iterators,
			createPagesIterator(key, repositoryURL, countPages),
		)
	}
	list, err := doCollect(
		iterators,
		func(repositoryURL string, pagesIterator *pagesiterator.PagesIterator, wg *sync.WaitGroup, cmp *concurrentmap.ConcurrentMap) {
			defer wg.Done()
			iterator := pagesIterator.Get(repositoryURL)
			issues := iterateHttpResponsesChannel(iterator.Responses)
			cmp.Set(repositoryURL, &issuePagesIterator{
				LastPage:   iterator.FirstPageNumber,
				URL:        repositoryURL,
				CountPages: iterator.LastPageNumber,
				Issues:     *issues,
			})
		},
	)
	if err != nil {
		return nil, err
	}
	return List(*list), nil
}

func findCountPages(repositoryURL string) (int64, error){
	response, err := requests.NewGET(repositoryURL, nil)
	if err != nil || response.StatusCode != http.StatusOK {
		return 0, err
	}
	list, err := deserializeIssuesList(response)
	if err != nil {
		return 0, err
	}
	response.Body.Close()
	return int64(list[0].Number / 100), nil
}

func doCollect(iterators []pagesiterator.Configurator, handle responsesHandler) (*concurrentmap.ConcurrentMap, error) {
	var (
		list = concurrentmap.New()
		wg   = new(sync.WaitGroup)
	)
	if len(iterators) == 0 {
		return nil, errors.New("Iterators not created. ")
	}
	pagesIterator := pagesiterator.NewPagesIterator(
		mapper.CoreRequestPerMinute,
		iterators...,
	)
	pagesIterator.DO()
	for repositoryURL := range pagesIterator.Iterators {
		wg.Add(1)
		go handle(repositoryURL, pagesIterator, wg, &list)
	}
	wg.Wait()
	return &list, nil
}

func iterateHttpResponsesChannel(responses chan *http.Response) *[]*mapper.Issue {
	var (
		mx     = new(sync.Mutex)
		wg     = new(sync.WaitGroup)
		issues = make([]*mapper.Issue, 0)
	)
	for response := range responses {
		wg.Add(1)
		go func(issues *[]*mapper.Issue, response *http.Response, wg *sync.WaitGroup, mx *sync.Mutex) {
			defer wg.Done()
			list, err := deserializeIssuesList(response)
			if err != nil {
				fmt.Println(runtimeinfo.Runtime(1), "; ERROR: ", err)
			} else {
				mx.Lock()
				*issues = append(*issues, list...)
				mx.Unlock()
			}
			response.Body.Close()
			return
		}(&issues, response, wg, mx)
	}
	wg.Wait()
	return &issues
}

func deserializeIssuesList(response *http.Response) ([]*mapper.Issue, error) {
	var list []*mapper.Issue
	err := requests.Deserialize(&list, response)
	if err != nil {
		return nil, err
	}
	return list, err
}

func createPagesIterator(key, url string, maxCountPages int64) pagesiterator.Configurator {
	return pagesiterator.NewConfiguration(
		pagesiterator.SetDefault(
			url,
			key,
			0,
		),
		pagesiterator.SetSizeBuffer(
			maxCountPages,
			100,
		),
	)
}
