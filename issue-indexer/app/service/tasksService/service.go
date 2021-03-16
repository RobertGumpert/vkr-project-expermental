package tasksService

import (
	concurrentMap "github.com/streamrail/concurrent-map"
	"issue-indexer/app/config"
	"issue-indexer/app/models/createTaskModel"
	"issue-indexer/app/models/updateTaskModel"
	"issue-indexer/app/repository"
	"issue-indexer/app/service/issuesComparator"
	"issue-indexer/pckg/runtimeinfo"
	"issue-indexer/pckg/task"
	"net/http"
	"sync"
)

type TasksService struct {
	MaxCountRunnableTasks int
	//
	mx                    *sync.Mutex
	config                *config.Config
	client                *http.Client
	comparator            *issuesComparator.IssuesComparator
	db                    repository.IRepositoriesStorage
	countNowRunnableTasks int
	tasks                 concurrentMap.ConcurrentMap
}

func NewTasksService(config *config.Config, db repository.IRepositoriesStorage) *TasksService {
	service := new(TasksService)
	service.client = new(http.Client)
	service.tasks = concurrentMap.New()
	service.config = config
	service.countNowRunnableTasks = 0
	service.mx = new(sync.Mutex)
	service.db = db
	service.MaxCountRunnableTasks = config.MaxCountRunnableTasks
	service.comparator = issuesComparator.NewComparator(
		config.MaxCountThreads,
		config.MinimumTextCompletenessThreshold,
		service.GettingResultFromComparator,
	)
	return service
}

func (service *TasksService) CreateTaskCompareIssuesInPairs(createTaskModel *createTaskModel.CreateTaskCompareIssuesInPairs) error {
	service.mx.Lock()
	defer service.mx.Unlock()
	taskState := &Task{
		Type:                        TypeTaskCompareIssuesInPairs,
		Key:                         createTaskModel.TaskKey,
		ExecutionStatus:             false,
		RunnableStatus:              false,
		ResultCompareFromComparator: nil,
		DeferStatus:                 false,
		taskContext:                 createTaskModel,
		sendResultToGate: &updateTaskModel.UpdateTaskCompareIssuesInPairs{
			TaskKey:                   createTaskModel.TaskKey,
			ComparableRepositoryID:    createTaskModel.ComparableRepositoryID,
			CompareWithRepositoriesID: createTaskModel.CompareWithRepositoriesID,
			CountNearestIssues:        0,
			NearestIssues:             make([]updateTaskModel.UpdateNearestIssues, 0),
		},
	}
	err := service.runTaskCompareIssuesInPairs(taskState)
	if err != nil {
		return err
	} else {
		service.tasks.Set(taskState.Key, taskState)
	}
	return nil
}

func (service *TasksService) GettingResultFromComparator(iTask task.ITask) {
	taskState := iTask.(*Task)
	if !service.tasks.Has(taskState.GetKey()) {
		runtimeinfo.LogError("GETTING RESULT UNREGISTERED TASK: key: ", taskState.GetKey(), ", type: ", taskState.Type)
		return
	}
	switch taskState.GetType() {
	case TypeTaskCompareIssuesInPairs:
		taskIsCompleted, err := service.updateTaskCompareIssuesInPairs(taskState)
		if err != nil {
			// TO DO:
		}
		if taskIsCompleted {
			//err := service.sendTaskCompareIssuesInPairs(taskState)
			//if err != nil {
			//	// TO DO:
			//}
			runtimeinfo.LogInfo("TASK WAS COMPLETED: key: ", taskState.GetKey())
			service.tasks.Pop(taskState.GetKey())
			service.runDeferTasks()
		}
		break
	default:
		runtimeinfo.LogError("GETTING RESULT UNREGISTERED TASK: key: ", taskState.GetKey(), ", type: ", taskState.Type)
	}
}

func (service *TasksService) runDeferTasks() {
	for taskInQueue := range service.tasks.IterBuffered() {
		taskState := taskInQueue.Val.(*Task)
		if taskState.DeferStatus && !taskState.RunnableStatus {
			err := taskState.runTaskTrigger(taskState)
			if err != nil {
				// TO DO:
			}
		}
	}
}
