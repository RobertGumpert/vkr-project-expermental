package appService

import (
	"github-gate/app/config"
	"github-gate/app/serivce/githubCollectorService"
	"github-gate/app/serivce/issueIndexerService"
	"github-gate/app/serivce/repositoryIndexerService"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/gotasker/tasker"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

type taskFacade struct {
	newRepositoryExistKeyword *taskNewRepositoryWithExistKeyWord
	newRepositoryNewKeyword   *taskNewRepositoryWithNewKeyword
	existRepository           *taskExistRepository
	//
	appService *AppService
}



func newTaskFacade(appService *AppService, db repository.IRepository, config *config.Config, engine *gin.Engine) *taskFacade {
	facade := new(taskFacade)
	//
	appService.serviceForCollector = githubCollectorService.NewCollectorService(
		db,
		config,
		engine,
	)
	appService.serviceForIssueIndexer = issueIndexerService.NewService(
		config,
		appService.channelResultsFromIssueIndexerService,
		engine,
	)
	appService.serviceForRepositoryIndexer = repositoryIndexerService.NewService(
		config,
		appService.channelResultsFromRepositoryIndexerService,
		engine,
	)
	appService.taskManager = tasker.NewManager(
		tasker.SetBaseOptions(
			100,
			facade.eventManageCompletedTasks,
		),
		tasker.SetRunByTimer(
			5*time.Second,
		),
	)
	//
	facade.appService = appService
	facade.newRepositoryExistKeyword = newTaskNewRepositoryWithExistKeyWord(appService)
	facade.newRepositoryNewKeyword = newTaskNewRepositoryWithNewKeyword(appService)
	facade.existRepository = newTaskExistRepository(appService)
	//
	go facade.scanChannelForCollectorService()
	go facade.scanChannelForIssueIndexerService()
	go facade.scanChannelForRepositoryIndexerService()
	//
	return facade
}

func (facade *taskFacade) GetNewRepositoryExistKeyword() *taskNewRepositoryWithExistKeyWord {
	return facade.newRepositoryExistKeyword
}

func (facade *taskFacade) GetNewRepositoryNewKeyword() *taskNewRepositoryWithNewKeyword {
	return facade.newRepositoryNewKeyword
}

func (facade *taskFacade) GetExistRepository() *taskExistRepository {
	return facade.existRepository
}

func (facade *taskFacade) scanChannelForCollectorService() {
	for task := range facade.appService.channelResultsFromCollectorService {
		repositories := task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		facade.appService.taskManager.SetUpdateForTask(
			task.GetKey(),
			repositories,
		)
	}
}

func (facade *taskFacade) scanChannelForIssueIndexerService() {
	for task := range facade.appService.channelResultsFromIssueIndexerService {
		facade.appService.taskManager.SetUpdateForTask(
			task.GetKey(),
			task.GetState().GetUpdateContext(),
		)
	}
}

func (facade *taskFacade) scanChannelForRepositoryIndexerService() {
	for task := range facade.appService.channelResultsFromRepositoryIndexerService {
		facade.appService.taskManager.SetUpdateForTask(
			task.GetKey(),
			task.GetState().GetUpdateContext(),
		)
	}
}

func (facade *taskFacade) eventManageCompletedTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	deleteTasks = make(map[string]struct{})
	switch task.GetType() {
	case TaskTypeNewRepositoryWithExistKeyword:
		deleteTasks = facade.newRepositoryExistKeyword.EventManageTasks(task)
		if len(deleteTasks) != 0 {
			if idDependent, trigger := task.IsDependent(); idDependent {
				repositories := trigger.GetState().GetUpdateContext().(*repositoryIndexerService.JsonSendFromIndexerReindexingForRepository)
				for id, distance := range repositories.Result.NearestRepositoriesID {
					log.Println("\t->Task Results : ", id, " = ", distance)
				}
			}
		}
		break
	case TaskTypeNewRepositoryWithNewKeyword:
		deleteTasks = facade.newRepositoryNewKeyword.EventManageTasks(task)
		if len(deleteTasks) != 0 {
			if idDependent, trigger := task.IsDependent(); idDependent {
				repositories := trigger.GetState().GetUpdateContext().(*repositoryIndexerService.JsonSendFromIndexerReindexingForRepository)
				for id, distance := range repositories.Result.NearestRepositoriesID {
					log.Println("\t->Task Results : ", id, " = ", distance)
				}
			}
		}
		break
	case TaskTypeExistRepository:
		deleteTasks = facade.existRepository.EventManageTasks(task)
		if len(deleteTasks) != 0 {
			if idDependent, trigger := task.IsDependent(); idDependent {
				repositories := trigger.GetState().GetUpdateContext().(*repositoryIndexerService.JsonSendFromIndexerReindexingForRepository)
				for id, distance := range repositories.Result.NearestRepositoriesID {
					log.Println("\t->Task Results : ", id, " = ", distance)
				}
			}
		}
		break
	}
	return deleteTasks
}
