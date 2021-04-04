package githubCollectorService

import (
	"errors"
	"github-gate/app/config"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/gotasker/tasker"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"net/http"
	"strings"
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

func (service *CollectorService) CreateGitHubApiURLForRepositories(repositories ...dataModel.RepositoryModel) (urls []string) {
	urls = make([]string, 0)
	for _, repo := range repositories {
		if strings.TrimSpace(repo.Owner) != "" && strings.TrimSpace(repo.Name) != "" {
			url := strings.Join([]string{
				gitHubApiAddress,
				"repos",
				repo.Owner,
				repo.Name,
			}, "/")
			urls = append(
				urls,
				url,
			)
		}
	}
	return urls
}

func (service *CollectorService) CreateTaskDescriptionRepositories(taskGateService itask.ITask, repositories ...dataModel.RepositoryModel) (err error) {
	if taskGateService == nil {
		return ErrorTaskIsNilPointer
	}
	if len(repositories) == 0 {
		return errors.New("Size of slice Data Models Repository is 0. ")
	}
	urls := service.CreateGitHubApiURLForRepositories(repositories...)
	if len(urls) == 0 {
		return errors.New("Failed to create url list. ")
	}
	constructor, err := service.createTaskRepositoriesDescriptions(
		taskGateService,
		urls...,
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

func (service *CollectorService) CreateTaskRepositoryIssues(taskGateService itask.ITask, repository dataModel.RepositoryModel) (err error) {
	if taskGateService == nil {
		return ErrorTaskIsNilPointer
	}
	urls := service.CreateGitHubApiURLForRepositories(repository)
	if len(urls) == 0 {
		return errors.New("Failed to create url list. ")
	}
	constructor, err := service.createTaskRepositoryIssues(
		taskGateService,
		urls[0],
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

func (service *CollectorService) CreateTaskRepositoriesDescriptionAndIssues(taskGateService itask.ITask, repositories ...dataModel.RepositoryModel) (err error) {
	if taskGateService == nil {
		return ErrorTaskIsNilPointer
	}
	if len(repositories) == 0 {
		return errors.New("Size of slice Data Models Repository is 0. ")
	}
	urls := service.CreateGitHubApiURLForRepositories(repositories...)
	if len(urls) == 0 {
		return errors.New("Failed to create url list. ")
	}
	triggers, err := service.createTaskRepositoriesDescriptionsAndIssues(
		taskGateService,
		urls...
	)
	if err != nil {
		return err
	}
	for _, trigger := range triggers {
		err := service.taskSteward.RunTask(trigger)
		if err != nil {
			runtimeinfo.LogError("RUNNING TRIGGER [", trigger.GetKey(), "] COMPLETED WITH ERROR: ", err)
		}
	}
	return nil
}
