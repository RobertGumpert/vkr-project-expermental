package appService

import (
	"github-gate/app/config"
	"github-gate/app/models/customFieldsModel"
	"github-gate/app/serivce/githubCollectorService"
	"github-gate/app/serivce/issueIndexerService"
	"github-gate/app/serivce/repositoryIndexerService"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/gotasker/tasker"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/requests"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
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
			facade.sendResultToApp(task, facade.appService.config.AppEndpoints.NearestRepositories)
		}
		break
	case TaskTypeNewRepositoryWithNewKeyword:
		deleteTasks = facade.newRepositoryNewKeyword.EventManageTasks(task)
		if len(deleteTasks) != 0 {
			facade.sendResultToApp(task, facade.appService.config.AppEndpoints.NearestRepositories)
		}
		break
	case TaskTypeExistRepository:
		deleteTasks = facade.existRepository.EventManageTasks(task)
		if len(deleteTasks) != 0 {
			facade.sendResultToApp(task, facade.appService.config.AppEndpoints.NearestRepositories)
		}
		break
	}
	return deleteTasks
}

func (facade *taskFacade) sendResultToApp(task itask.ITask, endpoint string) {
	if idDependent, trigger := task.IsDependent(); idDependent {
		userRequest := trigger.GetState().GetCustomFields().(*customFieldsModel.Model).GetContext().(JsonUserRequest)
		repositories := trigger.GetState().GetUpdateContext().(*repositoryIndexerService.JsonSendFromIndexerReindexingForRepository)
		jsonBody := &JsonSendToAppNearestRepositories{
			UserRequest:  userRequest,
			Repositories: repositories.Result.NearestRepositoriesID,
		}
		response, err := requests.POST(facade.appService.client, strings.Join([]string{
			facade.appService.config.AppAddress,
			endpoint,
		}, "/"), nil, jsonBody)
		if err != nil {
			runtimeinfo.LogError(err)
			return
		}
		if response.StatusCode != http.StatusOK {
			runtimeinfo.LogError("Status not 200. ")
			return
		}
	}
}
