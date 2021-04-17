package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/gotasker/tasker"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"time"
)

type taskFacade struct {
	byKeyword                                  *taskDownloadRepositoriesByKeyword
	byName                                     *taskDownloadRepositoriesByName
	repositoryAndRepositoriesContainingKeyword *taskDownloadRepositoryAndRepositoriesContainingKeyword
	channelErrors                              chan itask.IError
	service                                    *CollectorService
}

func newTaskFacade(service *CollectorService) *taskFacade {
	facade := new(taskFacade)
	facade.byKeyword = newTaskDownloadRepositoriesByKeyword(service)
	facade.byName = newTaskDownloadRepositoriesByName(service)
	facade.repositoryAndRepositoriesContainingKeyword = newTaskDownloadRepositoryAndRepositoriesContainingKeyword(service)
	facade.service = service
	service.taskManager = tasker.NewManager(
		tasker.SetBaseOptions(
			service.config.SizeQueueTasksForGithubCollectors,
			facade.eventManageCompletedTasks,
		),
		tasker.SetRunByTimer(
			1*time.Minute,
		),
	)
	facade.channelErrors = service.taskManager.GetChannelError()
	go facade.scanErrors()
	return facade
}

func (facade *taskFacade) GetRepositoryAndRepositoriesContainingKeyword() *taskDownloadRepositoryAndRepositoriesContainingKeyword {
	return facade.repositoryAndRepositoriesContainingKeyword
}

func (facade *taskFacade) GetByName() *taskDownloadRepositoriesByName {
	return facade.byName
}

func (facade *taskFacade) GetByKeyword() *taskDownloadRepositoriesByKeyword {
	return facade.byKeyword
}

func (facade *taskFacade) eventManageCompletedTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	switch task.GetType() {
	case TaskTypeDownloadCompositeByName:
		return facade.byName.EventManageTasks(task)
	case TaskTypeDownloadCompositeByKeyWord:
		return facade.byKeyword.EventManageTasks(task)
	case TaskTypeDownloadCompositeRepositoryAndRepositoriesContainingKeyWord:
		return facade.repositoryAndRepositoriesContainingKeyword.EventManageTasks(task)
	}
	return nil
}

func (facade *taskFacade) scanErrors() {
	for err := range facade.channelErrors {
		var (
			deleteKeys  = make(map[string]struct{})
			deleteTasks = make([]itask.ITask, 0)
		)
		//task, _ := err.GetTaskIfExist()
		//taskAppService := task.GetState().GetCustomFields().(itask.ITask)
		//taskAppService.GetState().SetError(err.GetError())
		runtimeinfo.LogError(err.GetError())
		deleteTasks = facade.service.taskManager.FindRunBanTriggers()
		deleteTasks = append(deleteTasks,  facade.service.taskManager.FindRunBanSimpleTasks()...)
		for _, task := range deleteTasks {
			deleteKeys[task.GetKey()] = struct{}{}
		}
		runtimeinfo.LogInfo("DELETE TASK WITH ERROR: ", deleteKeys)
		facade.service.taskManager.DeleteTasksByKeys(deleteKeys)
	}
}
