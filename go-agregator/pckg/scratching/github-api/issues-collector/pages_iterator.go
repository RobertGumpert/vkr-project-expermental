package issues_collector

import (
	"fmt"
	mapper "go-agregator/pckg/scratching/github-api/mapper"
	"strings"
)

type Option func(iterator *issuePagesIterator)
type Configurator func() *issuePagesIterator

type issuePagesIterator struct {
	LastPage   int64           `json:"last_page"`
	URL        string          `json:"url"`
	CountPages int64           `json:"count_pages"`
	Issues     []*mapper.Issue `json:"issues"`
}

func SetPage(page int64) Option {
	return func(iterator *issuePagesIterator) {
		iterator.LastPage = page
	}
}

func SetURL(url string) Option {
	return func(iterator *issuePagesIterator) {
		if !strings.Contains(url, "/issues?state=all") {
			url = fmt.Sprintf("%s/issues?state=all", url)
		}
		iterator.URL = url
	}
}

func SetCountPagesAll(url string) Option {
	return func(iterator *issuePagesIterator) {
		if !strings.Contains(url, "/issues?state=all") {
			url = fmt.Sprintf("%s/issues?state=all", url)
		}
		countPages, err := findCountPages(url)
		if err != nil {
			countPages = 0
		}
		iterator.CountPages = countPages
	}
}

func SetCountPages(countPages int64) Option {
	return func(iterator *issuePagesIterator) {
		iterator.CountPages = countPages
	}
}

func NewConfiguration(options ...Option) Configurator {
	return func() *issuePagesIterator {
		iterator := new(issuePagesIterator)
		for _, option := range options {
			option(iterator)
		}
		return iterator
	}
}