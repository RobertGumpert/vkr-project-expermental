package appService

import (
	"github.com/RobertGumpert/gosimstor"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"repository-indexer/app/config"
	"strings"
)



const (
	dictionaryDataModel          string = "dictionary"
	nearestRepositoriesDataModel string = "nearest"
	corpusDataModel              string = "corpus"
)

type AppService struct {
	config          *config.Config
	db              repository.IRepository
	localStorage    *gosimstor.Storage
	pathRootProject string
	maxCountThreads uint
	minIdf          uint
}

func NewAppService(pathRootProject string, config *config.Config, db repository.IRepository) (*AppService, error) {
	service := new(AppService)
	service.db = db
	service.config = config
	service.pathRootProject = strings.Join([]string{
		pathRootProject,
		"data",
		"storage",
	}, "/")
	localStorage, err := service.createLocalStorage()
	if err != nil {
		return nil, err
	}
	service.localStorage = localStorage
	return service, nil
}

func (service *AppService) createLocalStorage() (*gosimstor.Storage, error) {
	return gosimstor.NewStorage(
		gosimstor.NewFileProvider(
			dictionaryDataModel,
			service.pathRootProject,
			1,
			ToStringKeyWord,
			ToStringPositionKeyWord,
			FromStringToKeyWord,
			FromStringToPositionKeyWord,
		),
		//gosimstor.NewFileProvider(
		//	string(nearestRepositoriesDataModel),
		//	service.pathRootProject,
		//	3,
		//
		//	)
	)
}
