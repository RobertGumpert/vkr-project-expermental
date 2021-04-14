package issueIndexerService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (service *IndexerService) ConcatTheirRestHandlers(engine *gin.Engine) {
	updateTaskStateHandlers := engine.Group("/task/api/issueindexer/update")
	updateTaskStateHandlers.POST(
		"/compare/group",
		service.restHandlerUpdateCompareGroup,
	)
}

func (service *IndexerService) restHandlerUpdateCompareGroup(ctx *gin.Context) {
	state := new(JsonSendFromIndexerCompareGroup)
	if err := ctx.BindJSON(state); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	taskKey := state.ExecutionTaskStatus.TaskKey
	obj, exist := service.tasks.Get(taskKey)
	if exist {
		task := obj.(itask.ITask)
		go func(service *IndexerService, task itask.ITask) {
			task.GetState().SetError(state.ExecutionTaskStatus.Error)
			service.channelSendToAppService <- task
			service.tasks.Pop(task.GetKey())
			return
		}(service, task)
	}
}
