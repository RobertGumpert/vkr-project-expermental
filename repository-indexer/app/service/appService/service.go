package appService

import (
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"repository-indexer/app/config"
	"repository-indexer/app/service/indexerService"
)

type AppService struct {
	config                    *config.Config
	mainStorage, localStorage repository.IRepository
	minIdf                    uint
}

func NewAppService(config *config.Config, mainStorage, localStorage repository.IRepository) (*AppService, error) {
	service := new(AppService)
	service.localStorage = localStorage
	service.mainStorage = mainStorage
	service.config = config
	return service, nil
}

func (service *AppService) GetNearestRepositories(repositoryId uint) (dataModel.NearestRepositoriesJSON, error) {
	return service.localStorage.GetNearestRepositories(repositoryId)
}

func (service *AppService) reindexing() error {
	var (
		repositoriesIds     = make([]uint, 0)
		nearestRepositories = make([]dataModel.NearestRepositoriesJSON, 0)
		keyWords            = make([]dataModel.RepositoriesKeyWordsModel, 0)
	)
	models, err := service.mainStorage.GetAllRepositories()
	if err != nil {
		return err
	}
	results, err := indexerService.Indexing(models)
	if err != nil {
		return err
	}
	dictionary := results.GetDictionary()
	for item := range dictionary.IterBuffered() {
		keyWord := item.Key
		position := item.Val.(int64)
		keyWords = append(keyWords, dataModel.RepositoriesKeyWordsModel{
			KeyWord:  keyWord,
			Position: position,
		})
	}
	err = service.localStorage.RewriteAllKeyWords(keyWords)
	if err != nil {
		return err
	}
	resultsRepositoriesDistances := results.GetNearestRepositories()
	for _, repository := range resultsRepositoriesDistances {
		repositoriesIds = append(repositoriesIds, repository.GetRepositoryID())
		for id, distance := range repository.GetNearestRepositories() {

		}
		nearestRepositories = append(nearestRepositories, )
	}
	err = service.localStorage.RewriteAllNearestRepositories(repositoriesIds, nearestRepositories)
	if err != nil {
		return err
	}
}
