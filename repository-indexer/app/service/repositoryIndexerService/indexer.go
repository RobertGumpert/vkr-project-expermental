package repositoryIndexerService

import (
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	concurrentMap "github.com/streamrail/concurrent-map"
)

type repositoryIndexer struct {
	nearest []nearestRepository
	dictionary concurrentMap.ConcurrentMap
	minIdf                      uint
}

func IndexingIDF(models []dataModel.RepositoryModel, minIdf uint) (*repositoryIndexer, error) {
	indexer := new(repositoryIndexer)
	indexer.dictionary = concurrentMap.New()
	indexer.nearest = make([]nearestRepository, 0)
	indexer.minIdf = minIdf
	err := indexer.indexing(models)
	return indexer, err
}

func Indexing(models []dataModel.RepositoryModel) (*repositoryIndexer, error) {
	indexer := new(repositoryIndexer)
	indexer.dictionary = concurrentMap.New()
	indexer.nearest = make([]nearestRepository, 0)
	indexer.minIdf = 0
	err := indexer.indexing(models)
	return indexer, err
}

func (indexer *repositoryIndexer) GetNearestRepositories() []nearestRepository {
	return indexer.nearest
}

func (indexer *repositoryIndexer) GetDictionary() concurrentMap.ConcurrentMap {
	return indexer.dictionary
}
