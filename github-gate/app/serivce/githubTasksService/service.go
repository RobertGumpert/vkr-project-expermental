package githubTasksService

import (
	"github-gate/app/config"
	"github-gate/pckg/runtimeinfo"
	"github-gate/pckg/task"
	"net/http"
)

type QueueIsBusy func() bool
type RunTask func()

type GithubTasksService struct {
	config                    *config.Config
	client                    *http.Client
	tasksForCollectors        []*TaskForCollector
	tasksForCollectorsChannel chan *TaskForCollector
}

func NewGithubTasksService(config *config.Config, client *http.Client) *GithubTasksService {
	collectorTasks := make([]*TaskForCollector, 0)
	collectorTasksChannel := make(chan *TaskForCollector, config.CountTask)
	service := &GithubTasksService{
		config:                    config,
		client:                    client,
		tasksForCollectors:        collectorTasks,
		tasksForCollectorsChannel: collectorTasksChannel,
	}
	go service.scanChannelTasksForCollectors()
	return service
}

func (service *GithubTasksService) CreateTaskRepositoriesDescriptions(repositoriesUrls []string, iTask task.ITask) {

}

func (service *GithubTasksService) CreateTaskRepositoriesAndTheirIssues(repositoriesUrls []string, iTask task.ITask) (QueueIsBusy, RunTask) {
	var (
		queueIsBusy = func() bool {
			countTasks := 1 + len(repositoriesUrls)
			return service.queueIsBusy(countTasks)
		}
		runTask = func() {
			initializer, isDefer := service.newCollectorTaskRepositoriesDescriptionByURL(repositoriesUrls)
			dependent, _ := service.newListCollectorTasksRepositoriesIssues(repositoriesUrls)
			service.linkDependentCollectorTasks(initializer, dependent)
			if !isDefer {
				err, nonFreeCollectors := service.sendTaskToCollector(initializer)
				if err != nil && nonFreeCollectors {
					
				}
			}
		}
	)
	return queueIsBusy, nil
}

func (service *GithubTasksService) scanChannelTasksForCollectors() {
	for taskState := range service.tasksForCollectorsChannel {

	}
}


