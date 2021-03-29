package indexerService

import (
	"errors"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textDictionary"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textMetrics"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textVectorized"
	concurrentMap "github.com/streamrail/concurrent-map"
	"math"
	"strings"
)

func (results *indexingResults) indexingIDF(models []dataModel.RepositoryModel) error {
	var(
		corpus = results.createCorpus(models)
		dictionary concurrentMap.ConcurrentMap
		vectorOfWords [][]string
		err error
	)
	dictionary, vectorOfWords, err = results.createDictionary(corpus)
	bagOfWords, err := results.createBagOfWords(dictionary, vectorOfWords)
	if err != nil {
		return err
	}
	distances := results.calculateCosineDistance(bagOfWords)
	results.dictionary = dictionary
	for i := 0; i < len(distances); i++ {
		repository := nearestRepository{
			id:    models[i].ID,
			text:    corpus[i],
			nearest: make(map[uint]float64),
		}
		for j := 0; j < len(distances[i]); j++ {
			if math.IsNaN(distances[i][j]) {
				continue
			}
			if models[i].ID == models[j].ID {
				continue
			}
			if _, exist := repository.nearest[models[j].ID]; exist {
				continue
			} else {
				repository.nearest[models[j].ID] = distances[i][j]
			}
		}
		results.nearest = append(results.nearest, repository)
	}
	return nil
}

func (results *indexingResults) indexing(models []dataModel.RepositoryModel) error {
	var(
		corpus = results.createCorpus(models)
		dictionary concurrentMap.ConcurrentMap
		vectorOfWords [][]string
		err error
	)
	for i := 1; i < len(models) ;i++{
		results.minIdf = uint(i)
		dictionary, vectorOfWords, err = results.createDictionary(corpus)
		if err != nil {
			continue
		}
		if dictionary.Count() <= (len(models) / 2) {
			break
		}
	}
	bagOfWords, err := results.createBagOfWords(dictionary, vectorOfWords)
	if err != nil {
		return err
	}
	distances := results.calculateCosineDistance(bagOfWords)
	results.dictionary = dictionary
	for i := 0; i < len(distances); i++ {
		repository := nearestRepository{
			id:    models[i].ID,
			text:    corpus[i],
			nearest: make(map[uint]float64),
		}
		for j := 0; j < len(distances[i]); j++ {
			if math.IsNaN(distances[i][j]) {
				continue
			}
			if models[i].ID == models[j].ID {
				continue
			}
			if _, exist := repository.nearest[models[j].ID]; exist {
				continue
			} else {
				repository.nearest[models[j].ID] = distances[i][j]
			}
		}
		results.nearest = append(results.nearest, repository)
	}
	return nil
}

func (results *indexingResults) createCorpus(models []dataModel.RepositoryModel) []string {
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

func (results *indexingResults) createDictionary(corpus []string) (dictionary concurrentMap.ConcurrentMap, vectorsOfWords [][]string, err error) {
	dictionary, vectorsOfWords, count := textDictionary.IDFDictionary(
		corpus,
		int64(results.minIdf),
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

func (results *indexingResults) createBagOfWords(dictionary concurrentMap.ConcurrentMap, vectorsOfWords [][]string) (bagOfWords [][]float64, err error) {
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

func (results *indexingResults) calculateCosineDistance(bagOfWords [][]float64) (distances [][]float64) {
	return textMetrics.CosineDistance(bagOfWords, textPreprocessing.LinearMode)
}
