package repositoryComparator

import (
	"errors"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textDictionary"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textVectorized"
	concurrentMap "github.com/streamrail/concurrent-map"
	"repository-indexer/app/config"
	"strings"
)

type RepositoryComparator struct {
	config          *config.Config
	db              repository.IRepository
	maxCountThreads uint
	minIdf          uint
}

func NewRepositoryComparator(config *config.Config, db repository.IRepository) *RepositoryComparator {
	return &RepositoryComparator{config: config, db: db}
}

func (comparator *RepositoryComparator) readAllRepositories() ([]dataModel.RepositoryModel, error) {
	models, err := comparator.db.GetAllRepositories()
	if err != nil {
		return nil, err
	}
	if len(models) == 0 {
		return nil, errors.New("DOWNLOAD SIZE OF LIST MODELS EQUALS 0. ")
	}
	return models, err
}

func (comparator *RepositoryComparator) createCorpus(freshDownloadedModels []dataModel.RepositoryModel) []string {
	var (
		corpus = make([]string, 0)
	)
	for i := 0; i < len(freshDownloadedModels); i++ {
		repositoryModel := freshDownloadedModels[i]
		corpus = append(corpus, strings.Join([]string{
			repositoryModel.Description,
			strings.Join(repositoryModel.Topics, " "),
		}, " "))
	}
	return corpus
}

func (comparator *RepositoryComparator) createDictionary(corpus []string) (dictionary concurrentMap.ConcurrentMap, vectorsOfWords [][]string, err error) {
	dictionary, vectorsOfWords, count := textDictionary.IDFDictionary(
		corpus,
		int64(comparator.minIdf),
		textPreprocessing.LinearMode,
	)
	if count == 0 {
		return nil, nil, errors.New("COUNT FEATURES EQUALS 0. ")
	}
	if dictionary.Count() == 0 || len(vectorsOfWords) == 0 {
		return nil, nil, errors.New("DATA LEN. EQUALS 0. ")
	}
	if len(vectorsOfWords) != len(corpus) {
		return nil, nil, errors.New("LEN. VECTOR OF WORDS NOT EQUAL LEN. VECTOR OF CORPUS")
	}
	return dictionary, vectorsOfWords, nil
}

func (comparator *RepositoryComparator) vectorizedCorpus(dictionary concurrentMap.ConcurrentMap, vectorsOfWords [][]string) (bagOfWords [][]float64, err error) {
	bagOfWords = textVectorized.FrequencyVectorized(
		vectorsOfWords,
		dictionary,
		textPreprocessing.LinearMode,
	)
	if len(bagOfWords) != len(vectorsOfWords) {
		return nil, errors.New("LEN. BAG OF WORDS NOT EQUAL LEN. VECTOR OF WORDS. ")
	}
	return bagOfWords, nil
}


func (comparator *RepositoryComparator) createNearestRepositoriesHash(freshDownloadedModels []dataModel.RepositoryModel, bagOfWords [][]float64, dictionary concurrentMap.ConcurrentMap) {

}