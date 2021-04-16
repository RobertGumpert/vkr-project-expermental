package repositoryIndexerService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (service *IndexerService) ConcatTheirRestHandlers(engine *gin.Engine) {
	updateTaskStateHandlers := engine.Group("/task/api/repositoryindexer/update")
	updateTaskStateHandlers.POST(
		"/reindexing/for/repository",
		service.restHandlerReindexingForRepository,
	)
}

func (service *IndexerService) restHandlerReindexingForRepository(ctx *gin.Context) {
	state := new(JsonSendFromIndexerReindexingForRepository)
	if err := ctx.BindJSON(state); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	taskKey := state.ExecutionTaskStatus.TaskKey
	obj, exist := service.tasks.Get(taskKey)
	if exist {
		task := obj.(itask.ITask)
		task.GetState().SetUpdateContext(state)
		go func(service *IndexerService, task itask.ITask) {
			task.GetState().SetError(state.ExecutionTaskStatus.Error)
			service.channelSendToAppService <- task
			service.tasks.Pop(task.GetKey())
			return
		}(service, task)
	}
}
