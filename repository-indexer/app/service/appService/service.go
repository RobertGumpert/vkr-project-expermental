package appService

import (
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"net/http"
	"repository-indexer/app/config"
	"strings"
	"sync"
)

type AppService struct {
	config                    *config.Config
	mainStorage, localStorage repository.IRepository
	//
	chanResult         chan resultIndexing
	nowHaveRunningTask bool
	queue              []doReindexing
	mx                 *sync.Mutex
	client             *http.Client
}

func NewAppService(config *config.Config, mainStorage, localStorage repository.IRepository) (*AppService, error) {
	service := new(AppService)
	service.localStorage = localStorage
	service.mainStorage = mainStorage
	service.config = config
	service.queue = make([]doReindexing, 0)
	service.chanResult = make(chan resultIndexing)
	service.mx = new(sync.Mutex)
	service.client = new(http.Client)
	go service.scanChannel()
	return service, nil
}

func (service *AppService) RepositoryNearest(input *jsonInputNearestRepositoriesForRepository) (output *jsonOutputNearestRepositoriesForRepository) {
	var (
		haveTaskForReindexing = service.nowHaveRunningTask
	)
	if haveTaskForReindexing {
		return &jsonOutputNearestRepositoriesForRepository{
			NearestRepositories:  nil,
			DatabaseIsReindexing: haveTaskForReindexing,
		}
	} else {
		model, err := service.localStorage.GetNearestRepositories(input.RepositoryID)
		if err != nil {
			return &jsonOutputNearestRepositoriesForRepository{
				NearestRepositories:  nil,
				DatabaseIsReindexing: haveTaskForReindexing,
			}
		} else {
			return &jsonOutputNearestRepositoriesForRepository{
				NearestRepositories: []jsonNearestRepository{
					{
						RepositoryID:          input.RepositoryID,
						NearestRepositoriesID: model.Repositories,
					},
				},
				DatabaseIsReindexing: haveTaskForReindexing,
			}
		}
	}
}

func (service *AppService) WordIsExist(input *jsonInputWordIsExist) (output *jsonOutputWordIsExist) {
	var (
		haveTaskForReindexing = service.nowHaveRunningTask
	)
	model, err := service.localStorage.GetKeyWord(input.Word)
	if err != nil || model.ID == 0 {
		return &jsonOutputWordIsExist{
			WordIsExist:          false,
			DatabaseIsReindexing: haveTaskForReindexing,
		}
	}
	if model.KeyWord == input.Word {
		if haveTaskForReindexing {
			return &jsonOutputWordIsExist{
				WordIsExist:          true,
				DatabaseIsReindexing: haveTaskForReindexing,
			}
		} else {
			return &jsonOutputWordIsExist{
				WordIsExist:          true,
				DatabaseIsReindexing: haveTaskForReindexing,
			}
		}
	}
	return &jsonOutputWordIsExist{
		WordIsExist:          false,
		DatabaseIsReindexing: haveTaskForReindexing,
	}
}

func (service *AppService) QueueIsFilled() (isFilled bool) {
	if int64(len(service.queue)) >= service.config.MaximumSizeOfQueue {
		return true
	}
	return false
}

func (service *AppService) AddTask(jsonModel interface{}, taskType itask.Type) (err error) {
	if service.QueueIsFilled() {
		return gotasker.ErrorQueueIsFilled
	}
	service.mx.Lock()
	defer service.mx.Unlock()
	//
	switch taskType {
	case taskTypeReindexingForAll:
		service.addTaskReindexingForAll(jsonModel.(*jsonSendFromGateReindexingForAll))
		break
	case taskTypeReindexingForRepository:
		service.addTaskReindexingForRepository(jsonModel.(*jsonSendFromGateReindexingForRepository))
		break
	case taskTypeReindexingForGroupRepositories:
		service.addTaskReindexingForGroupRepositories(jsonModel.(*jsonSendFromGateReindexingForGroupRepositories))
		break
	}
	return
}

func (service *AppService) addTaskReindexingForRepository(jsonModel *jsonSendFromGateReindexingForRepository) {
	doReindex := service.getIndexerForRepository(jsonModel)
	service.queue = append(service.queue, doReindex)
	if !service.nowHaveRunningTask {
		service.nowHaveRunningTask = true
		runtimeinfo.LogInfo("RUN TASK: [", jsonModel.TaskKey, "]")
		go doReindex()
		return
	}
	runtimeinfo.LogInfo("TASK IS DEFER: [", jsonModel.TaskKey, "]")
	return
}

func (service *AppService) addTaskReindexingForAll(jsonModel *jsonSendFromGateReindexingForAll) {
	doReindex := service.getIndexerForAll(jsonModel)
	service.queue = append(service.queue, doReindex)
	if !service.nowHaveRunningTask {
		service.nowHaveRunningTask = true
		runtimeinfo.LogInfo("RUN TASK: [", jsonModel.TaskKey, "]")
		go doReindex()
		return
	}
	runtimeinfo.LogInfo("TASK IS DEFER: [", jsonModel.TaskKey, "]")
	return
}

func (service *AppService) addTaskReindexingForGroupRepositories(jsonModel *jsonSendFromGateReindexingForGroupRepositories) {
	doReindex := service.getIndexerForGroupRepositories(jsonModel)
	service.queue = append(service.queue, doReindex)
	if !service.nowHaveRunningTask {
		service.nowHaveRunningTask = true
		runtimeinfo.LogInfo("RUN TASK: [", jsonModel.TaskKey, "]")
		go doReindex()
		return
	}
	runtimeinfo.LogInfo("TASK IS DEFER: [", jsonModel.TaskKey, "]")
	return
}

func (service *AppService) scanChannel() {
	for result := range service.chanResult {
		service.sendTaskUpdateToGate(result)
		service.popFirstFromQueue()
		if len(service.queue) == 0 {
			service.nowHaveRunningTask = false
			continue
		} else {
			service.nowHaveRunningTask = true
			runtimeinfo.LogInfo("RUN TASK")
			go service.queue[0]()
		}
	}
}

func (service *AppService) popFirstFromQueue() {
	queue := make([]doReindexing, 0)
	for i := 1; i < len(service.queue); i++ {
		queue = append(queue, service.queue[i])
	}
	service.queue = queue
}

func (service *AppService) sendTaskUpdateToGate(result resultIndexing) {
	var (
		url string
		err error
	)
	switch result.taskType {
	case taskTypeReindexingForAll:
		url = strings.Join(
			[]string{
				service.config.GithubGateAddress,
				service.config.GithubGateEndpoints.SendResultTaskReindexingForAll,
			},
			"/",
		)
		err = result.jsonBody.(jsonSendToGateReindexingForAll).ExecutionTaskStatus.Error
		break
	case taskTypeReindexingForRepository:
		url = strings.Join(
			[]string{
				service.config.GithubGateAddress,
				service.config.GithubGateEndpoints.SendResultTaskReindexingForRepository,
			},
			"/",
		)
		err = result.jsonBody.(jsonSendToGateReindexingForRepository).ExecutionTaskStatus.Error
		break
	case taskTypeReindexingForGroupRepositories:
		url = strings.Join(
			[]string{
				service.config.GithubGateAddress,
				service.config.GithubGateEndpoints.SendResultTaskReindexingForGroupRepositories,
			},
			"/",
		)
		err = result.jsonBody.(jsonSendToGateReindexingForGroupRepositories).ExecutionTaskStatus.Error
		break
	}
	runtimeinfo.LogInfo("SEND TASK: [", result.taskKey, "] TO: [", url, "] WITH ERROR/NON ERROR: [", err, "]")
	runtimeinfo.LogInfo("SEND TASK: [", result.taskKey, "] TO: [", url, "] BODY: [", result.jsonBody, "]")

	//response, err := requests.POST(service.client, url, nil, result.jsonBody)
	//if err != nil {
	//	runtimeinfo.LogError(err)
	//}
	//if response.StatusCode != http.StatusOK {
	//	runtimeinfo.LogError("(REQ. -> TO GATE) STATUS NOT 200.")
	//}
}
