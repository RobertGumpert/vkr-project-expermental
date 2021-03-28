package repositoryIndexerService

import (
	"errors"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textDictionary"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textMetrics"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textVectorized"
	concurrentMap "github.com/streamrail/concurrent-map"
	"strings"
)

func (indexer *repositoryIndexer) indexing(models []dataModel.RepositoryModel) error {
	var(
		corpus = indexer.createCorpus(models)
		dictionary concurrentMap.ConcurrentMap
		vectorOfWords [][]string
		err error
	)
	for i := 1; i < len(models) ;i++{
		indexer.minIdf = uint(i)
		dictionary, vectorOfWords, err = indexer.createDictionary(corpus)
		if err != nil {
			continue
		}
		if dictionary.Count() <= len(models) {
			break
		}
	}
	bagOfWords, err := indexer.createBagOfWords(dictionary, vectorOfWords)
	if err != nil {
		return err
	}
	distances := indexer.calculateCosineDistance(bagOfWords)
	indexer.dictionary = dictionary
	for i := 0; i < len(distances); i++ {
		repository := nearestRepository{
			name:    models[i].Name,
			text:    corpus[i],
			nearest: make(map[string]float64),
		}
		for j := 0; j < len(distances[i]); j++ {
			if models[i].Name == models[j].Name {
				continue
			}
			if _, exist := repository.nearest[models[j].Name]; exist {
				continue
			} else {
				repository.nearest[models[j].Name] = distances[i][j]
			}
		}
		indexer.nearest = append(indexer.nearest, repository)
	}
	return nil
}

func (indexer *repositoryIndexer) createCorpus(models []dataModel.RepositoryModel) []string {
	var (
		corpus = make([]string, 0)
	)
	for i := 0; i < len(models); i++ {
		repositoryModel := models[i]
		corpus = append(corpus, strings.Join([]string{
			repositoryModel.Description,
			strings.Join(repositoryModel.Topics, " "),
		}, " "))
	}
	return corpus
}

func (indexer *repositoryIndexer) createDictionary(corpus []string) (dictionary concurrentMap.ConcurrentMap, vectorsOfWords [][]string, err error) {
	dictionary, vectorsOfWords, count := textDictionary.IDFDictionary(
		corpus,
		int64(indexer.minIdf),
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

func (indexer *repositoryIndexer) createBagOfWords(dictionary concurrentMap.ConcurrentMap, vectorsOfWords [][]string) (bagOfWords [][]float64, err error) {
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

func (indexer *repositoryIndexer) calculateCosineDistance(bagOfWords [][]float64) (distances [][]float64) {
	return textMetrics.CosineDistance(bagOfWords, textPreprocessing.LinearMode)
}
