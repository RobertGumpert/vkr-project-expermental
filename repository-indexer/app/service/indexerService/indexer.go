package indexerService

import (
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	concurrentMap "github.com/streamrail/concurrent-map"
)

type indexingResults struct {
	nearest []nearestRepository
	dictionary concurrentMap.ConcurrentMap
	minIdf                      uint
}

func (results *indexingResults) GetNearestRepositories() []nearestRepository {
	return results.nearest
}

func (results *indexingResults) GetDictionary() concurrentMap.ConcurrentMap {
	return results.dictionary
}

func IndexingIDF(models []dataModel.RepositoryModel, minIdf uint) (*indexingResults, error) {
	indexer := new(indexingResults)
	indexer.dictionary = concurrentMap.New()
	indexer.nearest = make([]nearestRepository, 0)
	indexer.minIdf = minIdf
	err := indexer.indexingIDF(models)
	return indexer, err
}

func Indexing(models []dataModel.RepositoryModel) (*indexingResults, error) {
	indexer := new(indexingResults)
	indexer.dictionary = concurrentMap.New()
	indexer.nearest = make([]nearestRepository, 0)
	indexer.minIdf = 0
	err := indexer.indexing(models)
	return indexer, err
}