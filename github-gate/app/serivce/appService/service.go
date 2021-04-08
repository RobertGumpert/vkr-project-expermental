package appService

import (
	"github-gate/app/config"
	"github-gate/app/serivce/githubCollectorService"
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/gotasker/tasker"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/gin-gonic/gin"
)

type AppService struct {
	db                          repository.IRepository
	config                      *config.Config
	taskManager                 itask.IManager
	collectorService            *githubCollectorService.CollectorService
	channelResultsFromCollector chan itask.ITask
}

func NewAppService(db repository.IRepository, config *config.Config, engine *gin.Engine) *AppService {
	service := new(AppService)
	service.db = db
	service.config = config
	service.taskManager = tasker.NewManager(
		tasker.SetBaseOptions(
			100,
			service.eventManageCompletedTasks,
		),
	)
	service.collectorService = githubCollectorService.NewCollectorService(
		db,
		config,
	)
	service.channelResultsFromCollector = make(chan itask.ITask)
	service.ConcatTheirRestHandlers(engine)
	service.collectorService.ConcatTheirRestHandlers(engine)
	go service.gettingResultFromCollectorService()
	return service
}

func (service *AppService) CreateApiTaskDownloadRepositoriesByNames(apiJsonModel *ApiJsonDownloadRepositoriesByName) (err error) {
	if isFilled := service.taskManager.QueueIsFilled(1); isFilled {
		return gotasker.ErrorQueueIsFilled
	}
	if len(apiJsonModel.Repositories) == 0 {
		return ErrorEmptyOrIncompleteJSONData
	}
	task, err := service.createTaskDownloadRepositoriesByName(
		ApiTaskDownloadRepositoryByName,
		apiJsonModel,
	)
	if err != nil {
		return err
	}
	err = service.taskManager.AddTaskAndTask(task)
	if err != nil {
		return gotasker.ErrorQueueIsFilled
	}
	return nil
}
