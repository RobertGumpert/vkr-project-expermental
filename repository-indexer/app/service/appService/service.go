package appService

import (
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/requests"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	concurrentMap "github.com/streamrail/concurrent-map"
	"net/http"
	"repository-indexer/app/config"
	"strings"
	"sync"
	"time"
)

type AppService struct {
	config                    *config.Config
	mainStorage, localStorage repository.IRepository
	reservedCopyKeywords      concurrentMap.ConcurrentMap
	//
	chanResult           chan resultIndexing
	databaseIsReindexing bool
	queue                []doReindexing
	mx                   *sync.Mutex
	client               *http.Client
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
		databaseIsReindexing = service.databaseIsReindexing
	)
	if databaseIsReindexing {
		return &jsonOutputNearestRepositoriesForRepository{
			NearestRepositories:  nil,
			DatabaseIsReindexing: databaseIsReindexing,
		}
	} else {
		model, err := service.localStorage.GetNearestRepositories(input.RepositoryID)
		if err != nil {
			return &jsonOutputNearestRepositoriesForRepository{
				NearestRepositories:  nil,
				DatabaseIsReindexing: databaseIsReindexing,
			}
		} else {
			return &jsonOutputNearestRepositoriesForRepository{
				NearestRepositories: []jsonNearestRepository{
					{
						RepositoryID:          input.RepositoryID,
						NearestRepositoriesID: model.Repositories,
					},
				},
				DatabaseIsReindexing: databaseIsReindexing,
			}
		}
	}
}

func (service *AppService) WordIsExist(input *jsonInputWordIsExist) (output *jsonOutputWordIsExist) {
	var (
		databaseIsReindexing = service.databaseIsReindexing
	)
	if databaseIsReindexing {
		return &jsonOutputWordIsExist{
			WordIsExist:          service.reservedCopyKeywords.Has(input.Word),
			DatabaseIsReindexing: databaseIsReindexing,
		}
	} else {
		model, err := service.localStorage.GetKeyWord(input.Word)
		if err != nil {
			return &jsonOutputWordIsExist{
				WordIsExist:          false,
				DatabaseIsReindexing: databaseIsReindexing,
			}
		} else {
			if model.KeyWord == input.Word {
				return &jsonOutputWordIsExist{
					WordIsExist:          true,
					DatabaseIsReindexing: databaseIsReindexing,
				}
			} else {
				return &jsonOutputWordIsExist{
					WordIsExist:          false,
					DatabaseIsReindexing: databaseIsReindexing,
				}
			}
		}
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
	if service.reservedCopyKeywords == nil {
		if err := service.createCopyKeywords(); err != nil {
			return err
		}
	}
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

func (service *AppService) createCopyKeywords() (err error) {
	keywords, err := service.localStorage.GetAllKeyWords()
	if err != nil {
		return err
	}
	service.reservedCopyKeywords = concurrentMap.New()
	for _, keyword := range keywords {
		service.reservedCopyKeywords.Set(keyword.KeyWord, 0)
	}
	return nil
}

func (service *AppService) addTaskReindexingForRepository(jsonModel *jsonSendFromGateReindexingForRepository) {
	doReindex := service.getIndexerForRepository(jsonModel)
	service.queue = append(service.queue, doReindex)
	if !service.databaseIsReindexing {
		service.databaseIsReindexing = true
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
	if !service.databaseIsReindexing {
		service.databaseIsReindexing = true
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
	if !service.databaseIsReindexing {
		service.databaseIsReindexing = true
		runtimeinfo.LogInfo("RUN TASK: [", jsonModel.TaskKey, "]")
		go doReindex()
		return
	}
	runtimeinfo.LogInfo("TASK IS DEFER: [", jsonModel.TaskKey, "]")
	return
}

func (service *AppService) scanChannel() {
	for result := range service.chanResult {
		runtimeinfo.LogInfo("TASK [", result.taskKey, "] FINISH.")
		time.Sleep(2*time.Second)
		service.sendTaskUpdateToGate(result)
		service.popFirstFromQueue()
		if len(service.queue) == 0 {
			service.databaseIsReindexing = false
			service.reservedCopyKeywords = nil
			continue
		} else {
			service.databaseIsReindexing = true
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
	//runtimeinfo.LogInfo("Send task ", result.taskKey, " = ",
	//	result.jsonBody, " to ", url, ", err : ", err)
	runtimeinfo.LogInfo(url, err)
	response, err := requests.POST(service.client, url, nil, result.jsonBody)
	if err != nil {
		runtimeinfo.LogError(err)
		return
	}
	if response.StatusCode != http.StatusOK {
		runtimeinfo.LogError("(REQ. -> TO GATE) STATUS NOT 200.")
	}
}
