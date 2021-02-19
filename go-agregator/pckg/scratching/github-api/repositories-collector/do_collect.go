package repositories_collector

import (
	"errors"
	"fmt"
	concurrentmap "github.com/streamrail/concurrent-map"
	"go-agregator/pckg/requests"
	"go-agregator/pckg/runtimeinfo"
	mapper "go-agregator/pckg/scratching/github-api/mapper"
	pagesiterator "go-agregator/pckg/scratching/github-api/pages-iterator"
	textPreprocessing "go-agregator/pckg/scratching/text-preprocessing"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

func CustomizableSearchCollect(configurators ...Configurator) (List, error) {
	if configurators == nil || len(configurators) == 0 {
		return nil, errors.New("Configurators not created. ")
	}
	var (
		wg                        = new(sync.WaitGroup)
		list                      = concurrentmap.New()
		configuratorsPageIterator = make([]pagesiterator.Configurator, 0)
	)
	for _, configurator := range configurators {
		repositoryIterator := configurator()
		if repositoryIterator == nil {
			continue
		}
		list.Set(repositoryIterator.Key, repositoryIterator)
		configuratorsPageIterator = append(configuratorsPageIterator, createPagesIterator(
			repositoryIterator.Key,
			repositoryIterator.URL,
			repositoryIterator.CountPages,
			0,
		))
	}
	pagesIterator := pagesiterator.NewPagesIterator(
		mapper.SearchRequestPerMinute,
		configuratorsPageIterator...,
	)
	pagesIterator.DO()
	for key := range pagesIterator.Iterators {
		wg.Add(1)
		go func(key string, list *concurrentmap.ConcurrentMap, pagesIterator *pagesiterator.PagesIterator, gwg *sync.WaitGroup) {
			defer gwg.Done()
			var (
				repositories = make([]*mapper.Repository, 0)
				mx           = new(sync.Mutex)
				wg           = new(sync.WaitGroup)
				iterator     = pagesIterator.Get(key)
			)
			for response := range iterator.Responses {
				wg.Add(1)
				go func(repositories *[]*mapper.Repository, response *http.Response, wg *sync.WaitGroup, mx *sync.Mutex) {
					defer wg.Done()
					//
					list, err := deserializeRepositoriesList(response)
					if err != nil {
						fmt.Println(runtimeinfo.Runtime(1), "; error:", err)
						return
					} else {
						for i := 0; i < len(list.Repositories); i++ {
							repository := list.Repositories[i]
							performDescriptionPreprocessing(repository)
						}
					}
					//
					mx.Lock()
					*repositories = append(*repositories, list.Repositories...)
					mx.Unlock()
					//
					return
				}(&repositories, response, wg, mx)
			}
			//
			wg.Wait()
			//
			item, _ := list.Get(key)
			repositoryPageIterator := item.(*repositoryPagesIterator)
			repositoryPageIterator.Repositories = repositories
			return
		}(key, &list, pagesIterator, wg)
	}
	wg.Wait()
	return List(list), nil
}

func Collect(url string, urls ...string) (List, error) {
	if len(urls) != 0 {
		urls = append(urls, url)
	} else {
		urls = make([]string, 0)
		urls = append(urls, url)
	}
	var (
		client  = new(http.Client)
		list    = concurrentmap.New()
		timeOut = time.Second * time.Duration(1)
	)
	for i := 0; i < len(urls); i++ {
		repositoryURL := urls[i]
		response, err := getRepositoryFromURL(
			repositoryURL,
			client,
		)
		if err != nil || response.StatusCode != http.StatusOK {
			fmt.Println(runtimeinfo.Runtime(1), " ERROR: status not 200 or have error: ", err)
			continue
		}
		repository, err := deserializeRepository(response)
		if err != nil {
			fmt.Println(runtimeinfo.Runtime(1), " ERROR: ", err)
			continue
		}
		performDescriptionPreprocessing(repository)
		list.Set(repositoryURL, &repositoryPagesIterator{
			Key:          repositoryURL,
			URL:          repositoryURL,
			Repositories: []*mapper.Repository{
				repository,
			},
		})
		fmt.Println(runtimeinfo.Runtime(1), fmt.Sprintf("; KEY[%s], OK Response code", repositoryURL))
		if i+1 == len(urls) {
			fmt.Println("->")
			break
		}
		fmt.Println(runtimeinfo.Runtime(1), "Time out ", timeOut)
		time.Sleep(timeOut)
	}
	return List(list), nil
}

func getRepositoryFromURL(repositoryURL string, client *http.Client) (*http.Response, error) {
	return requests.GET(
		client,
		repositoryURL,
		map[string]string{
			"Accept":        "application/vnd.github.mercy-preview+json",
			"Authorization": mapper.AUTHToken,
		},
	)
}

func performDescriptionPreprocessing(repository *mapper.Repository) {
	var (
		description string
	)
	if strings.TrimSpace((*repository).Description) == "" && len((*repository).Topics) == 0 {
		return
	}
	if len((*repository).Topics) != 0 && strings.TrimSpace((*repository).Description) != "" {
		description = strings.Join([]string{
			(*repository).Description,
			strings.Join((*repository).Topics, " "),
		}, " ")
	} else {
		if strings.TrimSpace((*repository).Description) != "" {
			description = (*repository).Description
		}
		if len((*repository).Topics) != 0 {
			description = strings.Join([]string{
				strings.Join((*repository).Topics, " "),
			}, " ")
		}
	}
	(*repository).DescriptionPreprossecing = textPreprocessing.NewTextPreprocessor(description).DO()
}

func createPagesIterator(key, url string, countPages, firstPage int64) pagesiterator.Configurator {
	return pagesiterator.NewConfiguration(
		pagesiterator.SetDefault(
			url,
			key,
			firstPage,
		),
		pagesiterator.SetSizeBuffer(
			countPages,
			100,
		),
		pagesiterator.SetHttpHeaders(
			map[string]string{
				"Accept": "application/vnd.github.mercy-preview+json",
			},
		),
	)
}

func findCountSearchPages(searchURL string) (int64, error) {
	response, err := requests.NewGET(searchURL, nil)
	defer response.Body.Close()
	if err != nil || response.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(response.Body)

		if err != nil {
			fmt.Println(runtimeinfo.Runtime(1), fmt.Sprintf("; ERROR [%s] %s: ", searchURL, err))
			return 0, err
		}
		bodyString := string(bodyBytes)
		fmt.Println(runtimeinfo.Runtime(1), fmt.Sprintf("; ERROR [%s] %s: ", searchURL, bodyString))
		return 0, err
	}
	list, err := deserializeRepositoriesList(response)
	if err != nil {
		return 0, err
	}
	return list.Count / 100, nil
}

func deserializeRepository(response *http.Response) (*mapper.Repository, error) {
	var repository *mapper.Repository
	err := requests.Deserialize(&repository, response)
	if err != nil {
		return nil, err
	}
	response.Body.Close()
	return repository, err
}

func deserializeRepositoriesList(response *http.Response) (*mapper.JSONRepositoriesList, error) {
	var list *mapper.JSONRepositoriesList
	err := requests.Deserialize(&list, response)
	if err != nil {
		return nil, err
	}
	response.Body.Close()
	return list, err
}
