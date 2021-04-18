package appService

import (
	"github-gate/app/config"
	"github-gate/app/serivce/githubCollectorService"
	"github-gate/app/serivce/issueIndexerService"
	"github-gate/app/serivce/repositoryIndexerService"
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/gin-gonic/gin"
)

type AppService struct {
	db          repository.IRepository
	config      *config.Config
	taskManager itask.IManager
	//
	channelResultsFromCollectorService         chan itask.ITask
	channelResultsFromIssueIndexerService      chan itask.ITask
	channelResultsFromRepositoryIndexerService chan itask.ITask
	//
	serviceForCollector         *githubCollectorService.CollectorService
	serviceForIssueIndexer      *issueIndexerService.IndexerService
	serviceForRepositoryIndexer *repositoryIndexerService.IndexerService
	//
	facade *taskFacade
}

func NewAppService(db repository.IRepository, config *config.Config, engine *gin.Engine) *AppService {
	service := new(AppService)
	service.db = db
	service.config = config
	service.channelResultsFromCollectorService = make(chan itask.ITask)
	service.channelResultsFromIssueIndexerService = make(chan itask.ITask)
	service.channelResultsFromRepositoryIndexerService = make(chan itask.ITask)
	service.facade = newTaskFacade(
		service,
		db,
		config,
		engine,
	)
	service.ConcatTheirRestHandlers(engine)
	return service
}


func (service *AppService) DownloadAndAnalyzeNewRepositoryWithExistKeyword(jsonModel *JsonNewRepositoryWithExistKeyword) (err error) {
	if isFilled := service.taskManager.QueueIsFilled(1); isFilled {
		return gotasker.ErrorQueueIsFilled
	}
	t := service.facade.GetNewRepositoryExistKeyword()
	task, err := t.CreateTask(jsonModel)
	if err != nil {
		return err
	}
	err = service.taskManager.AddTaskAndTask(task)
	if err != nil {
		return gotasker.ErrorQueueIsFilled
	}
	return nil
}

func (service *AppService) DownloadAndAnalyzeNewRepositoryWithNewKeyword(jsonModel *JsonNewRepositoryWithNewKeyword) (err error) {
	if isFilled := service.taskManager.QueueIsFilled(1); isFilled {
		return gotasker.ErrorQueueIsFilled
	}
	t := service.facade.GetNewRepositoryNewKeyword()
	task, err := t.CreateTask(jsonModel)
	if err != nil {
		return err
	}
	err = service.taskManager.AddTaskAndTask(task)
	if err != nil {
		return gotasker.ErrorQueueIsFilled
	}
	return nil
}
//
//func (service *AppService) CreateTaskCompositeNewRepositoryWithExistKeyWord(jsonModel *JsonNewRepositoryWithExistKeyword) (err error) {
//	if isFilled := service.taskManager.QueueIsFilled(3); isFilled {
//		return gotasker.ErrorQueueIsFilled
//	}
//	trigger, err := service.taskNewRepositoryWithExistWord.CreateTask(
//		jsonModel,
//		service.channelResultsFromCollectorService,
//	)
//	if err != nil {
//		return err
//	}
//	err = service.taskManager.AddTaskAndTask(trigger)
//	if err != nil {
//		return gotasker.ErrorQueueIsFilled
//	}
//	return nil
//}
//
//
//
//func (service *AppService) CreateTaskDownloadRepositoriesByNames(apiJsonModel *JsonNewRepositoryWithExistKeyword) (err error) {
//	if isFilled := service.taskManager.QueueIsFilled(1); isFilled {
//		return gotasker.ErrorQueueIsFilled
//	}
//	if len(apiJsonModel.Repositories) == 0 {
//		return ErrorEmptyOrIncompleteJSONData
//	}
//	task, err := service.createTaskDownloadRepositoriesByName(
//		TaskTypeDownloadRepositoryByName,
//		apiJsonModel,
//	)
//	if err != nil {
//		return err
//	}
//	err = service.taskManager.AddTaskAndTask(task)
//	if err != nil {
//		return gotasker.ErrorQueueIsFilled
//	}
//	return nil
//}
//
//func (service *AppService) CreateTaskDownloadRepositoriesByKeyWord(apiJsonModel *JsonSingleTaskDownloadRepositoriesByKeyWord) (err error) {
//	if isFilled := service.taskManager.QueueIsFilled(1); isFilled {
//		return gotasker.ErrorQueueIsFilled
//	}
//	if strings.TrimSpace(apiJsonModel.KeyWord) == "" {
//		return ErrorEmptyOrIncompleteJSONData
//	}
//	task, err := service.createTaskDownloadRepositoriesByKeyWord(
//		TaskTypeDownloadRepositoryByKeyWord,
//		apiJsonModel,
//	)
//	if err != nil {
//		return err
//	}
//	err = service.taskManager.AddTaskAndTask(task)
//	if err != nil {
//		return gotasker.ErrorQueueIsFilled
//	}
//	return nil
//}
//
//func (service *AppService) CreateTaskDownloadRepositoryAndRepositoriesByKeyWord(apiJsonModel *JsonSingleTaskDownloadRepositoryAndRepositoriesByKeyWord) (err error) {
//	if isFilled := service.taskManager.QueueIsFilled(1); isFilled {
//		return gotasker.ErrorQueueIsFilled
//	}
//	if strings.TrimSpace(apiJsonModel.KeyWord) == "" {
//		return ErrorEmptyOrIncompleteJSONData
//	}
//	task, err := service.createTaskDownloadRepositoryAndRepositoriesByKeyWord(
//		TaskTypeRepositoryAndRepositoriesByKeyWord,
//		apiJsonModel,
//	)
//	if err != nil {
//		return err
//	}
//	err = service.taskManager.AddTaskAndTask(task)
//	if err != nil {
//		return gotasker.ErrorQueueIsFilled
//	}
//	return nil
//}
