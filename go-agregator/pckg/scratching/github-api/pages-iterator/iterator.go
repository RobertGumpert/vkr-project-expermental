package pages_iterator

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-agregator/pckg/requests"
	"go-agregator/pckg/runtimeinfo"
	mapper "go-agregator/pckg/scratching/github-api/mapper"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Option func(iterator *iterator)
type Configurator func() *iterator

type iterator struct {
	IteratorKey         string `json:"iterator_key"`
	LastPageNumber      int64  `json:"last_page_number"`
	FirstPageNumber     int64  `json:"first_page_number"`
	URL                 string `json:"url"`
	CountElementsOnPage int64  `json:"count_elements_on_page"`
	//
	ApiRateLimit *time.Duration `json:"api_rate_limit"`
	UsedTimeOut  *time.Duration `json:"used_time_out"`
	//
	headers   map[string]string   `json:"-"`
	Responses chan *http.Response `json:"-"`
}

func SetDefault(url, iteratorKey string, firstPageNumber int64) Option {
	return func(iterator *iterator) {
		iterator.URL = url
		iterator.IteratorKey = iteratorKey
		iterator.FirstPageNumber = firstPageNumber
	}
}

func SetHttpHeaders(headers map[string]string) Option {
	return func(iterator *iterator) {
		iterator.headers = headers
	}
}

func SetSizeBuffer(countPages, countElementsOnPage int64) Option {
	return func(iterator *iterator) {
		iterator.LastPageNumber = countPages
		iterator.CountElementsOnPage = countElementsOnPage
		iterator.Responses = make(chan *http.Response, countPages)
	}
}

func NewConfiguration(options ...Option) Configurator {
	return func() *iterator {
		iterator := new(iterator)
		for _, option := range options {
			option(iterator)
		}
		if iterator.headers == nil {
			iterator.headers = make(map[string]string)
		}
		iterator.headers["Authorization"] = mapper.AUTHToken
		return iterator
	}
}

func (itr *iterator) iterate(wg *sync.WaitGroup) {
	var (
		reduceTimeOut = func(itr *iterator) {
			*itr.UsedTimeOut = *itr.UsedTimeOut - *itr.ApiRateLimit
		}
		client    = new(http.Client)
		doRequest = func(from int64, client *http.Client, iterator *iterator) error {
			url := strings.Join([]string{
				iterator.URL,
				fmt.Sprintf("&page=%d&per_page=%d", from, iterator.CountElementsOnPage),
			}, "")
			response, err := requests.GET(client, url, itr.headers)
			if err != nil {
				fmt.Println(runtimeinfo.Runtime(1), "; error: ", err)
				reduceTimeOut(itr)
				return err
			}
			if response.StatusCode != 200 {
				err := errors.New(fmt.Sprintf("; KEY[%s], ERROR response code is %d on page %d", iterator.IteratorKey, response.StatusCode, from))
				fmt.Println(runtimeinfo.Runtime(1), "; error: ", err)
				reduceTimeOut(itr)
				return err
			}
			iterator.Responses <- response
			fmt.Println(runtimeinfo.Runtime(1), fmt.Sprintf("; KEY[%s], OK Response code is %d on page %d. Buffer size / capacity = %d / %d", iterator.IteratorKey, response.StatusCode, from, len(iterator.Responses), cap(iterator.Responses)))
			return nil
		}
	)
	for ; itr.FirstPageNumber < itr.LastPageNumber; itr.FirstPageNumber++ {
		if err := doRequest(itr.FirstPageNumber, client, itr); err != nil {
			break
		}
		if itr.FirstPageNumber+1 < itr.LastPageNumber {
			fmt.Println(runtimeinfo.Runtime(1), "Time out ", itr.UsedTimeOut)
			time.Sleep(*itr.UsedTimeOut)
		}
	}
	//
	close(itr.Responses)
	reduceTimeOut(itr)
	wg.Done()
	//
	return
}

func (itr *iterator) Serialize() ([]byte, error) {
	bs, err := json.Marshal(itr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return bs, nil
}
