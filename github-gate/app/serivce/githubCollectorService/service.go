package githubCollectorService

import (
	"github-gate/app/config"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/gotasker/tasker"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"net/http"
	"time"
)

type CollectorService struct {
	config      *config.Config
	repository  repository.IRepository
	taskSteward *tasker.Steward
	client      *http.Client
}

func NewCollectorService(repository repository.IRepository, config *config.Config) *CollectorService {
	service := new(CollectorService)
	service.repository = repository
	service.config = config
	service.client = new(http.Client)
	service.taskSteward = tasker.NewSteward(
		config.SizeQueueTasksForGithubCollectors,
		time.Minute,
		service.eventManageCompletedTasks,
	)
	return service
}

func (service *CollectorService) CreateTaskDescriptionRepositories(taskGateService itask.ITask, urls []string) (err error) {
	constructor, err := service.createTaskRepositoriesDescriptions(
		taskGateService,
		urls,
	)
	if err != nil {
		return err
	}
	_, err = service.taskSteward.CreateTaskAndRun(constructor)
	if err != nil {
		return err
	}
	return nil
}

func (service *CollectorService) CreateTaskRepositoryIssues(taskGateService itask.ITask, url string) (err error) {
	constructor, err := service.createTaskRepositoryIssues(
		taskGateService,
		url,
	)
	if err != nil {
		return err
	}
	_, err = service.taskSteward.CreateTaskAndRun(constructor)
	if err != nil {
		return err
	}
	return nil
}
