package githubTasksService

import (
	"github-gate/app/config"
	concurrentMap "github.com/streamrail/concurrent-map"
	"net/http"
)

type GithubTasksService struct {
	config                    *config.Config
	client                    *http.Client
	tasksForCollectors        *concurrentMap.ConcurrentMap
	tasksForCollectorsChannel chan *TaskForCollector
}

func NewGithubTasksService(config *config.Config, client *http.Client) *GithubTasksService {
	collectorTasks := concurrentMap.New()
	collectorTasksChannel := make(chan *TaskForCollector, config.CountTask)
	service := &GithubTasksService{
		config:                    config,
		client:                    client,
		tasksForCollectors:        &collectorTasks,
		tasksForCollectorsChannel: collectorTasksChannel,
	}
	go service.scanTasksFromCollectors()
	return service
}

func (service *GithubTasksService) CreateTaskRepositoriesDescriptions() {
	
}

func (service *GithubTasksService) scanTasksFromCollectors() {
	for taskState := range service.tasksForCollectorsChannel {

	}
}
