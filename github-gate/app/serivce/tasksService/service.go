package tasksService

import (
	"github-gate/app/config"
	"github-gate/app/serivce/githubTasksService"
	"net/http"
)

type TaskService struct {
	client             *http.Client
	config             *config.Config
	gutHubTasksService *githubTasksService.GithubTasksService
}

func NewTaskService(config *config.Config) *TaskService {
	client := new(http.Client)
	gutHubTasksService := githubTasksService.NewGithubTasksService(config, client)
	service := &TaskService{
		client:             client,
		config:             config,
		gutHubTasksService: gutHubTasksService,
	}
	return service
}

func (service *TaskService) AddNewRepositories(repositoriesUrls []string) {

}