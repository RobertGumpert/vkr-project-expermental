package appService

import (
	"github-collector/app/config"
	"github-collector/app/service/githubApiService"
	"github-collector/pckg/runtimeinfo"
	"net/http"
)

type sendResponse func() error

type AppService struct {
	config            *config.Config
	client            *http.Client
	repeatedResponses []sendResponse
	GithubClient      *githubApiService.GithubClient
}

func NewAppService(config *config.Config) *AppService {
	service := &AppService{
		config: config,
		client: new(http.Client),
	}
	gitHubClient, err := githubApiService.NewGithubClient(service.config.GithubToken, service.config.CountTasks)
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	service.GithubClient = gitHubClient
	go service.repeatResponsesToGithubGate()
	return service
}

func (service *AppService) CreateTaskRepositoriesDescriptions(jsonModel *JsonCreateTaskRepositoriesDescriptions) (err error) {
	return service.createTaskRepositoriesDescriptions(jsonModel)
}

func (service *AppService) CreateTaskRepositoryIssues(jsonModel *JsonCreateTaskRepositoryIssues) (err error) {
	return service.createTaskRepositoryIssues(jsonModel)
}

func (service *AppService) CreateTaskRepositoriesByKeyWord(jsonModel *JsonCreateTaskRepositoriesByKeyWord) (err error) {
	return service.createTaskRepositoriesByKeyWord(jsonModel)
}
