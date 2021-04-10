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
	"time"
)

type CollectorService struct {
	config      *config.Config
	repository  repository.IRepository
	taskManager itask.IManager
	client      *http.Client
	channelErrors chan itask.IError
}

func NewCollectorService(repository repository.IRepository, config *config.Config) *CollectorService {
	service := new(CollectorService)
	service.repository = repository
	service.config = config
	service.client = new(http.Client)
	service.taskManager = tasker.NewManager(
		tasker.SetBaseOptions(
			config.SizeQueueTasksForGithubCollectors,
			service.eventManageCompletedTasks,
		),
		tasker.SetRunByTimer(
			1*time.Minute,
		),
	)
	service.channelErrors = service.taskManager.GetChannelError()
	go service.scanErrors()
	return service
}

func (service *CollectorService) scanErrors() {
	for err := range service.channelErrors {
		var(
			deleteKeys = make(map[string]struct{})
			deleteTasks = make([]itask.ITask, 0)
		)
		//task, _ := err.GetTaskIfExist()
		//taskAppService := task.GetState().GetCustomFields().(itask.ITask)
		//taskAppService.GetState().SetError(err.GetError())
		runtimeinfo.LogError(err.GetError())
		deleteTasks = service.taskManager.FindRunBanTriggers()
		deleteTasks = append(deleteTasks, service.taskManager.FindRunBanSimpleTasks()...)
		for _, task := range deleteTasks {
			deleteKeys[task.GetKey()] = struct{}{}
		}
		runtimeinfo.LogInfo("DELETE TASK WITH ERROR: ", deleteKeys)
		service.taskManager.DeleteTasksByKeys(deleteKeys)
	}
}


func (service *CollectorService) CreateSimpleTaskRepositoriesDescriptions(taskAppService itask.ITask, repositories ...dataModel.RepositoryModel) (err error) {
	if taskAppService == nil {
		return ErrorTaskIsNilPointer
	}
	if len(repositories) == 0 {
		return errors.New("Size of slice Data Models Repository is 0. ")
	}
	task, err := service.createTaskOnlyRepositoriesDescriptions(
		taskAppService,
		repositories...,
	)
	if err != nil {
		return err
	}
	err = service.taskManager.AddTaskAndTask(task)
	if err != nil {
		return err
	}
	return nil
}

func (service *CollectorService) CreateSimpleTaskRepositoryIssues(taskAppService itask.ITask, repository dataModel.RepositoryModel) (err error) {
	if taskAppService == nil {
		return ErrorTaskIsNilPointer
	}
	task, err := service.createTaskOnlyRepositoryIssues(
		taskAppService,
		repository,
	)
	if err != nil {
		return err
	}
	err = service.taskManager.AddTaskAndTask(task)
	if err != nil {
		return err
	}
	return nil
}

func (service *CollectorService) CreateTriggerTaskRepositoriesByName(taskAppService itask.ITask, repositories ...dataModel.RepositoryModel) (err error) {
	if taskAppService == nil {
		return ErrorTaskIsNilPointer
	}
	if len(repositories) == 0 {
		return errors.New("Size of slice Data Models Repository is 0. ")
	}
	triggers, err := service.createCompositeTaskSearchByName(
		taskAppService,
		repositories...
	)
	if err != nil {
		return err
	}
	for _, trigger := range triggers {
		err := service.taskManager.AddTaskAndTask(trigger)
		if err != nil {
			runtimeinfo.LogError("RUNNING TRIGGER [", trigger.GetKey(), "] COMPLETED WITH ERROR: ", err)
		}
	}
	return nil
}

func (service *CollectorService) CreateTriggerTaskRepositoriesByKeyWord(taskAppService itask.ITask, keyWord string) (err error) {
	if taskAppService == nil {
		return ErrorTaskIsNilPointer
	}
	trigger, err := service.createCompositeTaskSearchByKeyWord(
		taskAppService,
		keyWord,
	)
	if err != nil {
		return err
	}
	err = service.taskManager.AddTaskAndTask(trigger)
	if err != nil {
		runtimeinfo.LogError("RUNNING TRIGGER [", trigger.GetKey(), "] COMPLETED WITH ERROR: ", err)
	}
	return err
}

func (service *CollectorService) CreateTaskRepositoryAndRepositoriesByKeyWord(taskAppService itask.ITask, repository dataModel.RepositoryModel, keyWord string) (err error) {
	if taskAppService == nil {
		return ErrorTaskIsNilPointer
	}
	trigger, err := service.createTaskRepositoryAndRepositoriesContainingKeyWord(
		taskAppService,
		repository,
		keyWord,
	)
	if err != nil {
		return err
	}
	err = service.taskManager.AddTaskAndTask(trigger)
	if err != nil {
		runtimeinfo.LogError("RUNNING TRIGGER [", trigger.GetKey(), "] COMPLETED WITH ERROR: ", err)
	}
	return err
}