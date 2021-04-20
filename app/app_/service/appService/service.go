package appService

import (
	"app/app_/config"
	"app/app_/service/githubGateService"
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
	gateService       *githubGateService.Service
}

func NewAppService(db repository.IRepository, config *config.Config, engine *gin.Engine) *AppService {
	service := &AppService{db: db, config: config}
	service.ConcatTheirRestHandlers(engine)
	service.client = new(http.Client)
	return service
}

func (service *AppService) SendDeferResponseToClient(jsonModel *JsonFromGetNearestRepositories) {

}

func (service *AppService) FindNearestRepositories(jsonModel *JsonCreateTaskFindNearestRepositories) (err error) {
	if strings.TrimSpace(jsonModel.Name) == "" ||
		strings.TrimSpace(jsonModel.Owner) == "" ||
		strings.TrimSpace(jsonModel.Keyword) == "" ||
		strings.TrimSpace(jsonModel.Email) == "" {
		return errors.New("Empty JSON data. ")
	}
	if !service.isExistRepositoryAtGithub(jsonModel.Name, jsonModel.Owner) {
		return errors.New("Empty JSON data. ")
	}
	var (
		userRequest = githubGateService.JsonUserRequest{
			UserKeyword: jsonModel.Keyword,
			UserName:    jsonModel.Name,
			UserOwner:   jsonModel.Owner,
			UserEmail:   jsonModel.Email,
		}
		responseJsonBody = &JsonResultTaskFindNearestRepositories{
			Keyword: jsonModel.Keyword,
			Name:    jsonModel.Name,
			Owner:   jsonModel.Owner,
			Email:   jsonModel.Email,
			Top:     make([]JsonNearestRepository, 0),
		}
		repositoryModel dataModel.RepositoryModel
	)
	jsonWordIsExist, err := service.repositoryIndexer.WordIsExist(jsonModel.Keyword)
	if err != nil {
		return err
	}
	model, err := service.db.GetRepositoryByName(jsonModel.Name)
	if err == nil {
		//
		// Репозиторий существует в базе данных.
		//
		if jsonWordIsExist.WordIsExist == false {
			//
			// Если не существует слова,
			// считаем задачу как добавлеие нового
			// репощитория и нового слова.
			//
			err := service.gateService.CreateTaskNewRepositoryWithNewKeyword(
				jsonModel.Name,
				jsonModel.Owner,
				jsonModel.Keyword,
				userRequest,
			)
			if err != nil {
				return ErrorGateQueueIsFilled
			}
			return ErrorRequestReceivedLater
		}
		if jsonWordIsExist.WordIsExist == true {
			repositoryModel = model
			err := service.repositoryIsExist(jsonModel, repositoryModel, responseJsonBody, jsonWordIsExist.DatabaseIsReindexing)
			if err != nil {
				if err == ErrorRequestReceivedLater {
					err := service.gateService.CreateTaskExistRepositoryReindexing(
						jsonModel.Name,
						jsonModel.Owner,
						userRequest,
					)
					if err != nil {
						return ErrorGateQueueIsFilled
					}
					return ErrorRequestReceivedLater
				} else {
					return err
				}
			}
		}
	} else {
		if err == gorm.ErrRecordNotFound {
			//
			// Репозиторий не существует в базе данных.
			//
			if jsonWordIsExist.WordIsExist == true {
				//
				// Если существует слово,
				// считаем задачу как добавлеие нового
				// репозитория.
				//
				err := service.gateService.CreateTaskNewRepositoryWithExistKeyword(
					jsonModel.Name,
					jsonModel.Owner,
					userRequest,
				)
				if err != nil {
					return ErrorGateQueueIsFilled
				}
				return ErrorRequestReceivedLater
			}
			if jsonWordIsExist.WordIsExist == false {
				//
				// Если не существует слова,
				// считаем задачу как добавлеие нового
				// репозитория и нового слова.
				//
				err := service.gateService.CreateTaskNewRepositoryWithNewKeyword(
					jsonModel.Name,
					jsonModel.Owner,
					jsonModel.Keyword,
					userRequest,
				)
				if err != nil {
					return ErrorGateQueueIsFilled
				}
				return ErrorRequestReceivedLater
			}
		} else {
			return err
		}
	}
}

func (service *AppService) repositoryIsExist(
	jsonModel *JsonCreateTaskFindNearestRepositories,
	repositoryModel dataModel.RepositoryModel,
	responseJsonBody *JsonResultTaskFindNearestRepositories,
	databaseIsReindexing bool,
) (err error) {
	if databaseIsReindexing == true {
		//
		// В случае если база данных ключевых слов
		// перестраивается, то есть вероятность, того что появятся
		// новые соседи, что потребует для них посчитать расстояния
		// между ISSUE.
		//
		return ErrorRequestReceivedLater
	}
	if databaseIsReindexing == false {
		//
		// Найдем ближайших соседей.
		//
		jsonNearestRepositories, err := service.repositoryIndexer.GetNearestRepositories(repositoryModel.ID)
		if err != nil {
			return err
		}
		if jsonNearestRepositories.DatabaseIsReindexing == true {
			//
			// В случае если база данных ключевых слов
			// перестраивается, то есть вероятность, того что появятся
			// новые соседи, что потребует для них посчитать расстояния
			// между ISSUE.
			//
			return ErrorRequestReceivedLater
		}
		if jsonNearestRepositories.DatabaseIsReindexing == false {
			var (
				mapDistanceWithNearest = jsonNearestRepositories.NearestRepositories[0].NearestRepositoriesID
			)
			if len(jsonNearestRepositories.NearestRepositories) == 0 ||
				len(mapDistanceWithNearest) == 0 {
				//
				// В случае если ближайших соседей не нашлось
				// возвращаем пользователю ошибку.
				//
				return ErrorRepositoryDoesntNearestRepositories
			}
			err = service.fillTopNearestRepositories(
				repositoryModel.ID,
				responseJsonBody,
				mapDistanceWithNearest,
			)
			if err == nil {
				if len(responseJsonBody.Top) != len(mapDistanceWithNearest) {
					//
					// Если колчество пар, для которых был
					// проведен анализ сравнения ISSUE,
					// меньше чем количество соседей.
					//
					return ErrorRequestReceivedLater
				}
			} else {
				if err == gorm.ErrRecordNotFound {
					//
					// Если нет пар, для которых был
					// проведен анализ сравнения ISSUE,
					// для найденных соседей.
					//
					return ErrorRequestReceivedLater
				} else {
					return err
				}
			}
		}
	}
	return nil
}

func (service *AppService) taskNewRepositoryWithExistKeyWord(jsonModel *JsonCreateTaskFindNearestRepositories) {

}

func (service *AppService) fillTopNearestRepositories(repositoryId uint, responseJsonBody *JsonResultTaskFindNearestRepositories, mapDistanceWithNearest map[uint]float64) (err error) {
	intersectionModels, err := service.db.GetNumberIntersectionsForRepository(repositoryId)
	if err != nil {
		return err
	}
	for _, intersections := range intersectionModels {
		if distance, exist := mapDistanceWithNearest[intersections.ComparableRepositoryID]; exist {
			comparableModel, err := service.db.GetRepositoryByID(intersections.ComparableRepositoryID)
			if err != nil {
				continue
			}
			responseJsonBody.Top = append(
				responseJsonBody.Top,
				JsonNearestRepository{
					URL:                     comparableModel.URL,
					Name:                    comparableModel.Name,
					Owner:                   comparableModel.Owner,
					DescriptionDistance:     distance,
					NumberPairIntersections: intersections.NumberIntersections,
				},
			)
		}
	}
	return nil
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
