package appService

import (
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/gotasker/tasker"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"issue-indexer/app/config"
	"issue-indexer/app/service/issueCompator"
	"net/http"
	"strings"
	"time"
)

type AppService struct {
	taskManager   itask.IManager
	channelErrors chan itask.IError
	facade        *tasksFacade
	db            repository.IRepository
	config        *config.Config
	client        *http.Client
}

func NewAppService(db repository.IRepository, config *config.Config) *AppService {
	service := new(AppService)
	service.db = db
	service.config = config
	service.taskManager = tasker.NewManager(
		tasker.SetBaseOptions(
			int64(config.MaxCountRunnableTasks),
			service.eventManageTasks,
		),
		tasker.SetRunByTimer(
			10*time.Second,
		),
	)
	service.facade = newTasksFacade(
		service.taskManager,
		config,
		service.db,
	)
	service.client = new(http.Client)
	return service
}

func (service *AppService) CreateTaskCompareBeside(jsonModel *jsonSendFromCompareBeside) (err error) {
	if service.QueueIsFilled() {
		return gotasker.ErrorQueueIsFilled
	}
	taskFacade := service.facade.GetTaskCompareBesideRepository()
	task, err := taskFacade.CreateTask(
		jsonModel.TaskKey,
		jsonModel.RepositoryID,
		service.returnResultFromComparator,
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

func (service *AppService) CreateTaskCompareGroup(jsonModel *jsonSendFromGateCompareGroup) (err error) {
	if service.QueueIsFilled() {
		return gotasker.ErrorQueueIsFilled
	}
	taskFacade := service.facade.GetTaskCompareGroupRepositories()
	task, err := taskFacade.CreateTask(
		jsonModel.TaskKey,
		jsonModel.RepositoryID,
		jsonModel.ComparableRepositoriesID,
		service.returnResultFromComparator,
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

func (service *AppService) QueueIsFilled() (isFilled bool) {
	sizeQueue := service.taskManager.GetSizeQueue()
	if sizeQueue >= int64(service.config.MaxCountRunnableTasks) {
		return true
	}
	return false
}

func (service *AppService) eventManageTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	switch task.GetType() {
	case compareBesideRepository:
		deleteTasks = service.facade.GetTaskCompareBesideRepository().EventManageTasks(task)
		break
	case compareWithGroupRepositories:
		deleteTasks = service.facade.GetTaskCompareGroupRepositories().EventManageTasks(task)
		break
	}
	sendToGateCtx := task.GetState().GetCustomFields().(*sendToGateContext)
	service.returnResultToGate(sendToGateCtx)
	return deleteTasks
}

func (service *AppService) returnResultFromComparator(result *issueCompator.CompareResult) {
	task := result.GetIdentifier().(itask.ITask)
	task.GetState().SetCompleted(true)
	if err := result.GetErr(); err != nil {
		service.taskManager.SetRunBanInQueue(task)
		service.taskManager.SendErrorToErrorChannel(
			service.taskManager.CreateError(
				err,
				task.GetKey(),
				task,
			),
		)
	} else {
		service.taskManager.SetUpdateForTask(
			task.GetKey(),
			result,
		)
	}
}

func (service *AppService) scanErrors() {
	for err := range service.channelErrors {
		var (
			endpoint string
			task, _  = err.GetTaskIfExist()
			jsonBody = &jsonSendToGateCompareBeside{
				ExecutionTaskStatus: jsonExecutionTaskStatus{
					TaskKey:       task.GetKey(),
					TaskCompleted: true,
					Error:         err.GetError(),
				},
			}
			deleteKeys  = make(map[string]struct{})
			deleteTasks = make([]itask.ITask, 0)
		)
		switch task.GetType() {
		case compareWithGroupRepositories:
			endpoint = service.config.GithubGateEndpoints.SendResultTaskCompareGroup
			break
		case compareBesideRepository:
			endpoint = service.config.GithubGateEndpoints.SendResultTaskCompareBeside
			break
		}
		service.returnResultToGate(&sendToGateContext{
			endpoint: endpoint,
			taskKey:  err.GetTaskKey(),
			err:      err.GetError(),
			jsonBody: jsonBody,
		})
		runtimeinfo.LogError("TASK [", err.GetTaskKey(), "] UPDATE/RUN/COMPLETED WITH ERROR: ", err.GetError())
		deleteTasks = append(deleteTasks, service.taskManager.FindRunBanSimpleTasks()...)
		for _, task := range deleteTasks {
			deleteKeys[task.GetKey()] = struct{}{}
		}
		runtimeinfo.LogInfo("DELETE TASK WITH ERROR: ", deleteKeys)
		service.taskManager.DeleteTasksByKeys(deleteKeys)
	}
}

func (service *AppService) returnResultToGate(ctx *sendToGateContext) {
	url := strings.Join(
		[]string{
			service.config.GithubGateAddress,
			ctx.endpoint,
		},
		"/",
	)
	runtimeinfo.LogInfo("SEND TASK: [", ctx.taskKey, "] TO: [", url, "] WITH ERROR/NON ERROR: [", ctx.GetErr(), "]")
	//response, err := requests.POST(service.client, url, nil, ctx.jsonBody)
	//if err != nil {
	//	runtimeinfo.LogError(err)
	//}
	//if response.StatusCode != http.StatusOK {
	//	runtimeinfo.LogError("(REQ. -> TO GATE) STATUS NOT 200.")
	//}
}
