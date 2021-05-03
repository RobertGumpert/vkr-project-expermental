package appService

import (
	"app/app_/config"
	"app/app_/service/githubGateService"
	"app/app_/service/postService"
	"app/app_/service/repositoryIndexerService"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/requests"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"html/template"
	"net/http"
	"strings"
)

type AppService struct {
	db     repository.IRepository
	config *config.Config
	client *http.Client
	//
	deferResultEmailTemplate *template.Template
	//
	repositoryIndexer *repositoryIndexerService.Service
	gateService       *githubGateService.Service
	postAgent         *postService.Agent
}

func NewAppService(root string, db repository.IRepository, config *config.Config, engine *gin.Engine) *AppService {
	service := &AppService{db: db, config: config}
	service.ConcatTheirRestHandlers(root, engine)
	service.client = new(http.Client)
	service.repositoryIndexer = repositoryIndexerService.NewService(
		service.config,
		service.client,
	)
	service.gateService = githubGateService.NewService(
		service.client,
		service.config,
	)
	tmpl, err := template.ParseFiles(strings.Join([]string{
		root,
		"/data/assets/email-defer-message.html",
	}, ""))
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	service.deferResultEmailTemplate = tmpl
	postAgent, err := postService.NewTlsAgent(
		config.Posts[0].Boxes[0].Username,
		config.Posts[0].Boxes[0].Password,
		config.Posts[0].Boxes[0].Identity,
		config.Posts[0].TCPPort,
		config.Posts[0].Host,
		&tls.Config{ServerName: config.Posts[0].Host},
	)
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	service.postAgent = postAgent
	return service
}

func (service *AppService) SendDeferResponseToClient(jsonModel *JsonFromGetNearestRepositories) {
	responseJsonBody := &JsonResultTaskFindNearestRepositories{
		UserRequest: &JsonUserRequest{
			UserKeyword: jsonModel.UserRequest.UserKeyword,
			UserName:    jsonModel.UserRequest.UserName,
			UserOwner:   jsonModel.UserRequest.UserOwner,
			UserEmail:   jsonModel.UserRequest.UserEmail,
		},
		Top: make([]JsonNearestRepository, 0),
	}
	userRepository, err := service.db.GetRepositoryByName(jsonModel.UserRequest.UserName)
	if err != nil {
		runtimeinfo.LogError("DO NOT SEND LETTER TO CLIENT {", jsonModel.UserRequest.UserEmail, "} REPOS. { ", jsonModel.UserRequest.UserOwner, "/", jsonModel.UserRequest.UserName, " } WITH ERROR: {", err, "}")
		return
	}
	err = service.fillTopNearestRepositories(userRepository.ID, responseJsonBody, jsonModel.Repositories)
	if err != nil {
		runtimeinfo.LogError("DO NOT SEND LETTER TO CLIENT {", jsonModel.UserRequest.UserEmail, "} REPOS. { ", jsonModel.UserRequest.UserOwner, "/", jsonModel.UserRequest.UserName, " } WITH ERROR: {", err, "}")
		return
	}
	service.sortingTopRepositories(userRepository, responseJsonBody)
	hash, err := responseJsonBody.encodeHash()
	if err != nil {
		runtimeinfo.LogError("DO NOT SEND LETTER TO CLIENT {", jsonModel.UserRequest.UserEmail, "} REPOS. { ", jsonModel.UserRequest.UserOwner, "/", jsonModel.UserRequest.UserName, " } WITH ERROR: {", err, "}")
		return
	}
	url := strings.Join([]string{
		"get",
		hash,
	}, "/")
	msg := postService.NewMessage(service.postAgent.ClientBox(), jsonModel.UserRequest.UserEmail).Subject(
		"Пришли результаты поиска похожих репозиториев!",
	).DynamicHtml(
		service.deferResultEmailTemplate,
		struct {
			Name, Owner, URL string
		}{
			Name:  jsonModel.UserRequest.UserName,
			Owner: jsonModel.UserRequest.UserOwner,
			URL:   url,
		},
	)
	err = service.postAgent.SendLetter(msg.GetBytes(), msg.GetReceiver())
	if err != nil {
		runtimeinfo.LogError("DO NOT SEND LETTER TO CLIENT {", jsonModel.UserRequest.UserEmail, "} REPOS. { ", jsonModel.UserRequest.UserOwner, "/", jsonModel.UserRequest.UserName, " } WITH ERROR: {", err, "}")
	}
	return
}

func (service *AppService) GetNearestIssuesInPairNearestRepositories(mainRepositoryName, secondRepositoryName string) (responseJsonBody *JsonNearestIssues, err error) {
	if strings.TrimSpace(mainRepositoryName) == "" ||
		strings.TrimSpace(secondRepositoryName) == "" {
		return nil, errors.New("Empty JSON data. ")
	}
	mainRepository, err := service.db.GetRepositoryByName(mainRepositoryName)
	if err != nil {
		return nil, err
	}
	secondRepository, err := service.db.GetRepositoryByName(secondRepositoryName)
	if err != nil {
		return nil, err
	}
	nearestIssuesList, err := service.db.GetNearestIssuesForPairRepositories(mainRepository.ID, secondRepository.ID)
	if err != nil {
		return nil, err
	}
	responseJsonBody = &JsonNearestIssues{
		UserRepositoryName:       mainRepository.Name,
		ComparableRepositoryName: secondRepository.Name,
		Top:                      make([]JsonNearestIssue, 0),
	}
	for next := 0; next < len(nearestIssuesList); next++ {
		nearestIssues := nearestIssuesList[next]
		mainRepositoryIssue, err := service.db.GetIssueByID(nearestIssues.IssueID)
		if err != nil {
			continue
		}
		secondRepositoryIssue, err := service.db.GetIssueByID(nearestIssues.NearestIssueID)
		if err != nil {
			continue
		}
		jsonModelNearestIssues := JsonNearestIssue{
			UserRepositoryName:  mainRepository.Name,
			UserRepositoryTitle: mainRepositoryIssue.Title,
			UserRepositoryURL:   strings.ReplaceAll(mainRepositoryIssue.URL, "api.github.com/repos", "github.com"),
			//
			ComparableRepositoryName:  secondRepository.Name,
			ComparableRepositoryTitle: secondRepositoryIssue.Title,
			ComparableRepositoryURL:   strings.ReplaceAll(secondRepositoryIssue.URL, "api.github.com/repos", "github.com"),
			//
			Rank:                int64(nearestIssues.Rank * 100),
			TitleCosine:         int64(nearestIssues.TitleCosineDistance),
			BodyCosine:          int64(nearestIssues.BodyCosineDistance),
			TopicsIntersections: nearestIssues.Intersections,
		}
		responseJsonBody.Top = append(responseJsonBody.Top, jsonModelNearestIssues)
	}
	responseJsonBody.makeTop()
	return responseJsonBody, nil
}

func (service *AppService) FindNearestRepositories(jsonModel *JsonCreateTaskFindNearestRepositories) (responseJsonBody *JsonResultTaskFindNearestRepositories, err error) {
	if strings.TrimSpace(jsonModel.Name) == "" ||
		strings.TrimSpace(jsonModel.Owner) == "" ||
		strings.TrimSpace(jsonModel.Keyword) == "" ||
		strings.TrimSpace(jsonModel.Email) == "" {
		return nil, errors.New("Empty JSON data. ")
	}
	if !service.isExistRepositoryAtGithub(jsonModel.Name, jsonModel.Owner) {
		return nil, errors.New("Empty JSON data. ")
	}
	var (
		userRequest = githubGateService.JsonUserRequest{
			UserKeyword: jsonModel.Keyword,
			UserName:    jsonModel.Name,
			UserOwner:   jsonModel.Owner,
			UserEmail:   jsonModel.Email,
		}
		repositoryModel dataModel.RepositoryModel
	)
	responseJsonBody = &JsonResultTaskFindNearestRepositories{
		UserRequest: &JsonUserRequest{
			UserKeyword: jsonModel.Keyword,
			UserName:    jsonModel.Name,
			UserOwner:   jsonModel.Owner,
			UserEmail:   jsonModel.Email,
		},
		Top: make([]JsonNearestRepository, 0),
	}
	jsonWordIsExist, err := service.repositoryIndexer.WordIsExist(jsonModel.Keyword)
	if err != nil {
		return nil, err
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
				return nil, ErrorGateQueueIsFilled
			}
			responseJsonBody.Defer = true
			return responseJsonBody, ErrorRequestReceivedLater
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
						return nil, ErrorGateQueueIsFilled
					}
					responseJsonBody.Defer = true
					return responseJsonBody, ErrorRequestReceivedLater
				} else {
					return nil, err
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
					return nil, ErrorGateQueueIsFilled
				}
				responseJsonBody.Defer = true
				return responseJsonBody, ErrorRequestReceivedLater
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
					return nil, ErrorGateQueueIsFilled
				}
				responseJsonBody.Defer = true
				return responseJsonBody, ErrorRequestReceivedLater
			}
		} else {
			return nil, err
		}
	}
	responseJsonBody.Defer = false
	service.sortingTopRepositories(repositoryModel, responseJsonBody)
	return responseJsonBody, nil
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

func (service *AppService) sortingTopRepositories(userRepository dataModel.RepositoryModel, responseJsonBody *JsonResultTaskFindNearestRepositories) {
	responseJsonBody.makeTop()
	responseJsonBody.UserRepository = &JsonUserRepository{
		URL:         userRepository.URL,
		Name:        userRepository.Name,
		Owner:       userRepository.Owner,
		Topics:      userRepository.Topics,
		Description: userRepository.Description,
	}
	var (
		makeTopicsToMap = func(topics []string) map[string]bool {
			mp := make(map[string]bool)
			for _, topic := range topics {
				mp[topic] = true
			}
			return mp
		}
		makeDescriptionToMap = func(description string) map[string]bool {
			mp := make(map[string]bool)
			words := strings.Split(description, " ")
			for _, word := range words {
				word = strings.TrimSpace(word)
				mp[word] = true
			}
			return mp
		}
		mapIntersections = func(mpUserRepository, mpNearestRepository map[string]bool) []string {
			intersections := make([]string, 0)
			for topic, _ := range mpNearestRepository {
				if _, exist := mpUserRepository[topic]; exist {
					intersections = append(
						intersections,
						topic,
					)
				}
			}
			return intersections
		}
		userRepositoryTopicsMap      = makeTopicsToMap(userRepository.Topics)
		userRepositoryDescriptionMap = makeDescriptionToMap(userRepository.Description)
	)
	for next := 0; next < len(responseJsonBody.Top); next++ {
		nearest := &responseJsonBody.Top[next]
		nearestRepositoryTopicsMap := makeTopicsToMap(nearest.Topics)
		nearestRepositoryDescriptionMap := makeDescriptionToMap(nearest.Description)
		intersectionsTopics := mapIntersections(userRepositoryTopicsMap, nearestRepositoryTopicsMap)
		intersectionsDescriptions := mapIntersections(userRepositoryDescriptionMap, nearestRepositoryDescriptionMap)
		nearest.TopicsIntersections = intersectionsTopics
		nearest.DescriptionIntersections = intersectionsDescriptions
	}
	return
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
			if comparableModel.ID == repositoryId {
				continue
			}
			responseJsonBody.Top = append(
				responseJsonBody.Top,
				JsonNearestRepository{
					URL:   comparableModel.URL,
					Name:  comparableModel.Name,
					Owner: comparableModel.Owner,
					//
					Topics:      comparableModel.Topics,
					Description: comparableModel.Description,
					//
					DescriptionDistance:     distance,
					NumberPairIntersections: fmt.Sprintf("%.6f", intersections.NumberIntersections),
					//
					RepositoryCountIssues: intersections.RepositoryCountIssues,
					CountNearestPairs:     intersections.CountNearestPairs,
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
