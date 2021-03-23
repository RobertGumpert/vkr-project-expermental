package hashService

import (
	"errors"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textDictionary"
	"github.com/RobertGumpert/vkr-pckg/textPreprocessing/textVectorized"
	concurrentMap "github.com/streamrail/concurrent-map"
	"strings"
)

func (storage *HashStorageService) Reindex(newStorage *HashStorageService) error {
	newStorage.dictionary = concurrentMap.New()
	newStorage.links = concurrentMap.New()
	newStorage.bagOfWords = nil
	newStorage.db = storage.db
	newStorage.minIdf = storage.minIdf
	//
	models, err := newStorage.download()
	if err != nil {
		return err
	}
	corpus := newStorage.createCorpus(models)
	newStorage.linkNameWithPositionIndexInBagOfWords(corpus, models)
	dictionary, vectorOfWords, err := newStorage.createDictionary(corpus)
	if err != nil {
		return err
	}
	bagOfWords, err := newStorage.vectorizedCorpus(dictionary, vectorOfWords)
	if err != nil {
		return err
	}
}

func (storage *HashStorageService) download() ([]dataModel.RepositoryModel, error) {
	models, err := storage.db.GetAllRepositories()
	if err != nil {
		return nil, err
	}
	if len(models) == 0 {
		return nil, errors.New("DOWNLOAD SIZE OF LIST MODELS EQUALS 0. ")
	}
	return models, err
}

func (storage *HashStorageService) createCorpus(models []dataModel.RepositoryModel) []string {
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

func (storage *HashStorageService) linkNameWithPositionIndexInBagOfWords(corpus []string, models []dataModel.RepositoryModel) {
	for index := 0; index < len(corpus); index++ {
		if !strings.Contains(corpus[index], models[index].Description) {
			storage.links.Set(models[index].Name, -1)
		} else {
			storage.links.Set(models[index].Name, index)
		}
	}
	return
}

func (storage *HashStorageService) linkNameWithKeyWords(corpus []string, models []dataModel.RepositoryModel) {
	for index := 0; index < len(corpus); index++ {
		if !strings.Contains(corpus[index], models[index].Description) {
			storage.links.Set(models[index].Name, -1)
		} else {
			storage.links.Set(models[index].Name, index)
		}
	}
	return
}

func (storage *HashStorageService) createDictionary(corpus []string) (dictionary concurrentMap.ConcurrentMap, vectorsOfWords [][]string, err error) {
	dictionary, vectorsOfWords, count := textDictionary.IDFDictionary(
		corpus,
		int64(storage.minIdf),
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

func (storage *HashStorageService) vectorizedCorpus(dictionary concurrentMap.ConcurrentMap, vectorsOfWords [][]string) (bagOfWords [][]float64, err error) {
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
