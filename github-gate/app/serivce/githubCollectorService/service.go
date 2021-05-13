package githubCollectorService

import (
	"github-gate/app/config"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CollectorService struct {
	config      *config.Config
	repository  repository.IRepository
	taskManager itask.IManager
	client      *http.Client
	//
	facade *taskFacade
}

func NewCollectorService(repository repository.IRepository, config *config.Config, engine *gin.Engine) *CollectorService {
	service := new(CollectorService)
	service.repository = repository
	service.config = config
	service.client = new(http.Client)
	if engine != nil {
		service.ConcatTheirRestHandlers(engine)
	}
	//
	service.facade = newTaskFacade(service)
	return service
}

func (service *CollectorService) CreateTaskRepositoriesByKeyword(taskAppService itask.ITask, keyWord string) (err error) {
	t := service.facade.GetByKeyword()
	task, err := t.CreateTask(
		taskAppService,
		keyWord,
	)
	if err != nil {
		return err
	}
	return service.taskManager.AddTaskAndTask(task)
}

func (service *CollectorService) CreateTaskRepositoriesByName(taskAppService itask.ITask, repositoryDataModels ...dataModel.RepositoryModel) (err error) {
	t := service.facade.GetByName()
	tasks, err := t.CreateTask(
		taskAppService,
		repositoryDataModels...,
	)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		err := service.taskManager.AddTaskAndTask(task)
		if err != nil {
			runtimeinfo.LogError("RUNNING TASK [", task.GetKey(), "] COMPLETED WITH ERROR: ", err)
		}
	}
	return nil
}

func (service *CollectorService) CreateTaskRepositoryAndRepositoriesContainingKeyword(taskAppService itask.ITask, repositoryDataModel dataModel.RepositoryModel, keyWord string) (err error) {
	t := service.facade.GetRepositoryAndRepositoriesContainingKeyword()
	task, err := t.CreateTask(
		taskAppService,
		repositoryDataModel,
		keyWord,
	)
	if err != nil {
		return err
	}
	return service.taskManager.AddTaskAndTask(task)
}

//
//func (service *CollectorService) CreateSimpleTaskRepositoriesDescriptions(taskAppService itask.ITask, repositories ...dataModel.RepositoryModel) (err error) {
//	if taskAppService == nil {
//		return ErrorTaskIsNilPointer
//	}
//	if len(repositories) == 0 {
//		return errors.New("Size of slice Data Models Repository is 0. ")
//	}
//	task, err := service.createTaskOnlyRepositoriesDescriptions(
//		taskAppService,
//		repositories...,
//	)
//	if err != nil {
//		return err
//	}
//	err = service.taskManager.AddTaskAndTask(task)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (service *CollectorService) CreateSimpleTaskRepositoryIssues(taskAppService itask.ITask, repository dataModel.RepositoryModel) (err error) {
//	if taskAppService == nil {
//		return ErrorTaskIsNilPointer
//	}
//	task, err := service.createTaskOnlyRepositoryIssues(
//		taskAppService,
//		repository,
//	)
//	if err != nil {
//		return err
//	}
//	err = service.taskManager.AddTaskAndTask(task)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (service *CollectorService) CreateTriggerTaskRepositoriesByName(taskAppService itask.ITask, repositories ...dataModel.RepositoryModel) (err error) {
//	if taskAppService == nil {
//		return ErrorTaskIsNilPointer
//	}
//	if len(repositories) == 0 {
//		return errors.New("Size of slice Data Models Repository is 0. ")
//	}
//	triggers, err := service.createCompositeTaskSearchByName(
//		taskAppService,
//		repositories...
//	)
//	if err != nil {
//		return err
//	}
//	for _, trigger := range triggers {
//		err := service.taskManager.AddTaskAndTask(trigger)
//		if err != nil {
//			runtimeinfo.LogError("RUNNING TRIGGER [", trigger.GetKey(), "] COMPLETED WITH ERROR: ", err)
//		}
//	}
//	return nil
//}
//
//func (service *CollectorService) CreateTriggerTaskRepositoriesByKeyWord(taskAppService itask.ITask, keyWord string) (err error) {
//	if taskAppService == nil {
//		return ErrorTaskIsNilPointer
//	}
//	trigger, err := service.createCompositeTaskSearchByKeyWord(
//		taskAppService,
//		keyWord,
//	)
//	if err != nil {
//		return err
//	}
//	err = service.taskManager.AddTaskAndTask(trigger)
//	if err != nil {
//		runtimeinfo.LogError("RUNNING TRIGGER [", trigger.GetKey(), "] COMPLETED WITH ERROR: ", err)
//	}
//	return err
//}
//
//func (service *CollectorService) CreateTaskRepositoryAndRepositoriesByKeyWord(taskAppService itask.ITask, repository dataModel.RepositoryModel, keyWord string) (err error) {
//	if taskAppService == nil {
//		return ErrorTaskIsNilPointer
//	}
//	trigger, err := service.createTaskRepositoryAndRepositoriesContainingKeyWord(
//		taskAppService,
//		repository,
//		keyWord,
//	)
//	if err != nil {
//		return err
//	}
//	err = service.taskManager.AddTaskAndTask(trigger)
//	if err != nil {
//		runtimeinfo.LogError("RUNNING TRIGGER [", trigger.GetKey(), "] COMPLETED WITH ERROR: ", err)
//	}
//	return err
//}
