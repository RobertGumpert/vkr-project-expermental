package appService

import (
	"app/app_/config"
	"app/app_/service/repositoryIndexerService"
	"errors"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/requests"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type AppService struct {
	db     repository.IRepository
	config *config.Config
	client *http.Client
	//
	repositoryIndexer *repositoryIndexerService.Service
}

func NewAppService(db repository.IRepository, config *config.Config, engine *gin.Engine) *AppService {
	service := &AppService{db: db, config: config}
	service.ConcatTheirRestHandlers(engine)
	service.client = new(http.Client)
	return service
}

func (service *AppService) FindNearestRepositories(jsonModel *JsonTaskFindNearestRepositories) (err error) {
	if strings.TrimSpace(jsonModel.Name) == "" ||
		strings.TrimSpace(jsonModel.Owner) == "" ||
		strings.TrimSpace(jsonModel.Keyword) == "" ||
		strings.TrimSpace(jsonModel.Email) == "" {
		return errors.New("Empty JSON data. ")
	}
	if !service.isExistRepositoryAtGithub(jsonModel.Name, jsonModel.Owner) {
		return errors.New("Empty JSON data. ")
	}
	repositoryDataModel, intersectionsDataModel, err := service.isExistRepositoryAtDatabase(jsonModel.Name)
	if err != nil && err == gorm.ErrRecordNotFound {
		if repositoryDataModel.ID == 0 {
			//
			//
			//
		}
	} else {
		return err
	}
	wordIsExist, err := service.repositoryIndexer.WordIsExist(jsonModel.Keyword)
	if err != nil {
		return err
	}
	if wordIsExist.DatabaseIsReindexing {

	}
}

func (service *AppService) isExistRepositoryAtGithub(name, owner string) (isExist bool) {
	var (
		url = strings.Join(
			[]string{
				"https://github.com",
				owner,
				name,
			},
			"/",
		)
	)
	response, err := requests.GET(
		service.client,
		url,
		nil,
	)
	if err != nil {
		return false
	}
	if response.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func (service *AppService) isExistRepositoryAtDatabase(name string) (
	repositoryDataModel dataModel.RepositoryModel,
	intersectionsDataModel []dataModel.NumberIssueIntersectionsModel,
	err error,
) {

	repositoryDataModel, err = service.db.GetRepositoryByName(name)
	if err != nil {
		return dataModel.RepositoryModel{
			Model: gorm.Model{ID: 0},
		}, nil, err
	}
	intersectionsDataModel, err = service.db.GetNumberIntersectionsForRepository(repositoryDataModel.ID)
	if err != nil {
		return repositoryDataModel, intersectionsDataModel, err
	}
	return repositoryDataModel, intersectionsDataModel, nil
}
