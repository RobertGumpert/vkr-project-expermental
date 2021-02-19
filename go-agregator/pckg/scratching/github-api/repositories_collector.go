package github_api

import (
	"encoding/json"
	"fmt"
	concurrent_map "github.com/streamrail/concurrent-map"
	"go-agregator/pckg/runtimeinfo"
	pages_iterator "go-agregator/pckg/scratching/github-api/pages-iterator"
	text_preprocessing "go-agregator/pckg/scratching/text-preprocessing"
	"net/http"
	"strings"
	"sync"
	"time"
)


type List *concurrent_map.ConcurrentMap

type repositoriesCollector struct {
	RootURL string                        `json:"-"`
	KeyList *concurrent_map.ConcurrentMap `json:"key_list"`
	mx      *sync.Mutex
}

func NewRepositoriesCollector() *repositoriesCollector {
	return &repositoriesCollector{
		RootURL: "",
	}
}

func (collector *repositoriesCollector) LanguagesCollect(maxCountPages, countElementsOnPage int64, languages ...Language) *repositoriesCollector {
	if languages == nil || len(languages) == 0 {
		return collector
	}
	collector.RootURL = "https://api.github.com/search/repositories?q=language:%s"
	mapRepositoriesByLanguages := concurrent_map.New()
	collector.KeyList = &mapRepositoriesByLanguages
	iteratorConstructors := make([]pages_iterator.Configurator, 0)
	for _, language := range languages {
		iteratorConstructor := pages_iterator.NewConfiguration(
			pages_iterator.SetDefault(
				fmt.Sprintf(collector.RootURL, string(language)),
				string(language),
				0,
			),
			pages_iterator.SetSizeBuffer(
				maxCountPages,
				countElementsOnPage,
			),
			pages_iterator.SetHttpHeaders(
				map[string]string{
					"Accept": "application/vnd.github.mercy-preview+json",
				},
			),
		)
		iteratorConstructors = append(iteratorConstructors, iteratorConstructor)
	}
	pagesIterator := pages_iterator.NewPagesIterator(
		10,
		iteratorConstructors...,
	)
	pagesIterator.DO()
	wg := new(sync.WaitGroup)
	for language := range pagesIterator.Iterators {
		wg.Add(1)
		go func(language string, pagesIterator *pages_iterator.PagesIterator, wg *sync.WaitGroup) {
			defer wg.Done()
			iterator := pagesIterator.Get(language)
			repos := collector.unmarshalHttpResponsesFromChannel(iterator.Responses)
			collector.KeyList.Set(language, repos)
		}(language, pagesIterator, wg)
	}
	wg.Wait()
	return collector
}

func (collector *repositoriesCollector) Serialize() ([]byte, string, error) {
	bs, err := json.Marshal(collector)
	if err != nil {
		fmt.Println(err)
		return nil, "", err
	}
	return bs, string(bs), nil
}

func (collector *repositoriesCollector) unmarshalHttpResponsesFromChannel(responses chan *http.Response) []*repository {
	var (
		repos                      = make([]*repository, 0)
	)
	start := time.Now()
	for response := range responses {
		list := new(repositoriesList)
		err := json.NewDecoder(response.Body).Decode(list)
		if err != nil {
			fmt.Println(runtimeinfo.Runtime(1), "; error:", err)
		} else {
			repos = append(repos, list.Repositories...)
		}
		response.Body.Close()
	}
	for _, repo := range repos {
		var description string
		if strings.TrimSpace(repo.Description) == "" && len(repo.Topics) == 0 {
			continue
		}
		if len(repo.Topics) != 0 && strings.TrimSpace(repo.Description) != "" {
			description = strings.Join([]string{
				repo.Description,
				strings.Join(repo.Topics, " "),
			}, " ")
		} else {
			if strings.TrimSpace(repo.Description) != "" {
				description = repo.Description
			}
			if len(repo.Topics) != 0 {
				description = strings.Join([]string{
					strings.Join(repo.Topics, " "),
				}, " ")
			}
		}
		repo.DescriptionPreprossecing = text_preprocessing.NewTextPreprocessor(description).DO()
	}
	duration := time.Since(start)
	fmt.Println(duration)
	return repos
}
