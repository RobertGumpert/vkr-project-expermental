package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker/itask"
)

type taskDownloadRepositoryAndRepositoriesContainingKeyword struct {
	service *CollectorService
}

func (t *taskDownloadRepositoryAndRepositoriesContainingKeyword) CreateTask() {

}

func (t *taskDownloadRepositoryAndRepositoriesContainingKeyword) EventManageTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	panic("implement me")
}

func (t *taskDownloadRepositoryAndRepositoriesContainingKeyword) EventUpdateTaskState(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	panic("implement me")
}

func (t *taskDownloadRepositoryAndRepositoriesContainingKeyword) EventRunTask(task itask.ITask) (doTaskAsDefer, sendToErrorChannel bool, err error) {
	panic("implement me")
}


