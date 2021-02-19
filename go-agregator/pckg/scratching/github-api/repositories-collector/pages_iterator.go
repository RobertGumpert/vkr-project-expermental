package repositories_collector

import (
	"fmt"
	mapper "go-agregator/pckg/scratching/github-api/mapper"
	"strconv"
	"strings"
)

const (
	baseSearchURL           = "https://api.github.com/search/repositories?q="
	searchTemplateLanguage  = "language:%s"
	searchTemplateStars     = "stars:%s"
	searchTemplateFollowers = "followers:%s"
	searchTemplateTopic     = "topic:%s"
	searchTemplateIn        = "%s in:description"
)

type Option func(iterator *repositoryPagesIterator)
type Configurator func() *repositoryPagesIterator

type repositoryPagesIterator struct {
	Key              string               `json:"key"`
	IsFindCount      bool                 `json:"is_find_count"`
	SearchParameters map[string]string    `json:"search_parameters"`
	LastPage         int64                `json:"last_page"`
	URL              string               `json:"url"`
	CountPages       int64                `json:"count_pages"`
	Repositories     []*mapper.Repository `json:"repositories"`
}

//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//--------------------------------------------From page-----------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------

func SetPage(page int64) Option {
	return func(iterator *repositoryPagesIterator) {
		iterator.LastPage = page
	}
}

//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//--------------------------------------------Topic---------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------

func SetTopic(topic string) Option {
	return func(iterator *repositoryPagesIterator) {
		if _, exist := iterator.SearchParameters[searchTemplateTopic]; !exist {
			iterator.SearchParameters[searchTemplateTopic] = fmt.Sprintf(
				searchTemplateTopic,
				topic,
			)
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//--------------------------------------------Language------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------

func SetLanguage(language mapper.Language) Option {
	return func(iterator *repositoryPagesIterator) {
		if _, exist := iterator.SearchParameters[searchTemplateLanguage]; !exist {
			iterator.SearchParameters[searchTemplateLanguage] = fmt.Sprintf(
				searchTemplateLanguage,
				string(language),
			)
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//--------------------------------------------Stars---------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------

func SetStarsOnlyValue(value int) Option {
	return func(iterator *repositoryPagesIterator) {
		if _, exist := iterator.SearchParameters[searchTemplateStars]; !exist {
			iterator.SearchParameters[searchTemplateStars] = fmt.Sprintf(
				searchTemplateStars,
				strconv.Itoa(value),
			)
		}
	}
}

func SetStarsInSegmentValues(from, to int) Option {
	return func(iterator *repositoryPagesIterator) {
		if _, exist := iterator.SearchParameters[searchTemplateStars]; !exist {
			iterator.SearchParameters[searchTemplateStars] = fmt.Sprintf(
				searchTemplateStars,
				fmt.Sprintf(
					"%s..%s",
					strconv.Itoa(from),
					strconv.Itoa(to),
				),
			)
		}
	}
}

func SetStarsFromValue(from int) Option {
	return func(iterator *repositoryPagesIterator) {
		if _, exist := iterator.SearchParameters[searchTemplateStars]; !exist {
			iterator.SearchParameters[searchTemplateStars] = fmt.Sprintf(
				searchTemplateStars,
				fmt.Sprintf(
					">=%s",
					strconv.Itoa(from),
				),
			)
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//--------------------------------------------Followers-----------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------

func SetFollowersOnlyValue(value int) Option {
	return func(iterator *repositoryPagesIterator) {
		if _, exist := iterator.SearchParameters[searchTemplateFollowers]; !exist {
			iterator.SearchParameters[searchTemplateFollowers] = fmt.Sprintf(
				searchTemplateFollowers,
				strconv.Itoa(value),
			)
		}
	}
}

func SetFollowersInSegmentValues(from, to int) Option {
	return func(iterator *repositoryPagesIterator) {
		if _, exist := iterator.SearchParameters[searchTemplateFollowers]; !exist {
			iterator.SearchParameters[searchTemplateFollowers] = fmt.Sprintf(
				searchTemplateFollowers,
				fmt.Sprintf(
					"%s..%s",
					strconv.Itoa(from),
					strconv.Itoa(to),
				),
			)
		}
	}
}

func SetFollowersFromValue(from int) Option {
	return func(iterator *repositoryPagesIterator) {
		if _, exist := iterator.SearchParameters[searchTemplateFollowers]; !exist {
			iterator.SearchParameters[searchTemplateFollowers] = fmt.Sprintf(
				searchTemplateFollowers,
				fmt.Sprintf(
					">=%s",
					strconv.Itoa(from),
				),
			)
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//--------------------------------------------Count Pages---------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------

func SetCountPagesAll() Option {
	return func(iterator *repositoryPagesIterator) {
		iterator.IsFindCount = true
	}
}

func SetCountPages(countPages int64) Option {
	return func(iterator *repositoryPagesIterator) {
		iterator.CountPages = countPages
	}
}

//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//--------------------------------------------IN---------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------

func SetTextIn(in string) Option {
	return func(iterator *repositoryPagesIterator) {
		if _, exist := iterator.SearchParameters[searchTemplateIn]; !exist {
			//iterator.SearchParameters[searchTemplateIn] = fmt.Sprintf(
			//	searchTemplateIn,
			//	in,
			//)
			iterator.SearchParameters[searchTemplateIn] = strings.ReplaceAll(in, " ", "%20")
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//--------------------------------------------Configurator--------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------------------------------------

func NewConfiguration(key string, options ...Option) Configurator {
	return func() *repositoryPagesIterator {
		iterator := new(repositoryPagesIterator)
		iterator.Key = key
		iterator.IsFindCount = false
		iterator.SearchParameters = make(map[string]string, 0)
		//
		for _, option := range options {
			option(iterator)
		}
		//
		if len(iterator.SearchParameters) == 0 {
			return nil
		}
		url := baseSearchURL
		for _, param := range iterator.SearchParameters {
			url += fmt.Sprintf("%s+", param)
		}
		url = url[:len(url)-1]
		iterator.URL = url
		//
		if iterator.IsFindCount {
			if countPages, err := findCountSearchPages(iterator.URL); err == nil {
				iterator.CountPages = countPages
			} else {
				return nil
			}
		}
		//
		return iterator
	}
}
