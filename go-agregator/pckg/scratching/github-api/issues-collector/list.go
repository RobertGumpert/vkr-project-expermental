package issues_collector

import (
	"encoding/json"
	"fmt"
	concurrentmap "github.com/streamrail/concurrent-map"
	mapper "go-agregator/pckg/scratching/github-api/mapper"
)

type List concurrentmap.ConcurrentMap

func (list List) GetKeys() []string {
	return concurrentmap.ConcurrentMap(list).Keys()
}

func (list List) Get(key string) *issuePagesIterator {
	element, ok := concurrentmap.ConcurrentMap(list).Get(key)
	if !ok {
		return nil
	}
	iterator := element.(*issuePagesIterator)
	return iterator
}

func (list List) GetIssues(key string) []*mapper.Issue {
	element, ok := concurrentmap.ConcurrentMap(list).Get(key)
	if !ok {
		return nil
	}
	iterator := element.(*issuePagesIterator)
	return iterator.Issues
}

func (list List) Serialize() ([]byte, string, error) {
	bs, err := json.Marshal(concurrentmap.ConcurrentMap(list))
	if err != nil {
		fmt.Println(err)
		return nil, "", err
	}
	return bs, string(bs), nil
}

func DeserializeToList(bt []byte) List {
	var (
		mp map[string]*issuePagesIterator
		list = concurrentmap.New()
	)
	err := json.Unmarshal(bt, &mp)
	if err != nil {
		return nil
	}
	for key, value := range mp {
		list.Set(key, value)
	}
	return List(list)
}
