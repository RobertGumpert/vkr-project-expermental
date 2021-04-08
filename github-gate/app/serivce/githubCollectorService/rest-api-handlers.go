package githubCollectorService

import (
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (service *CollectorService) ConcatTheirRestHandlers(engine *gin.Engine) {
	updateTaskStateHandlers := engine.Group("/task/api/collector/update")
	updateTaskStateHandlers.POST(
		"/repositories/descriptions",
		service.restHandlerUpdateDescriptionsRepositories,
	)
	updateTaskStateHandlers.POST(
		"/repository/issues",
		service.restHandlerUpdateRepositoryIssues,
	)
}

func (service *CollectorService) restHandlerUpdateDescriptionsRepositories(context *gin.Context) {
	state := new(jsonSendFromCollectorDescriptionsRepositories)
	if err := context.BindJSON(state); err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		context.AbortWithStatus(http.StatusLocked)
		return
	}
	service.taskManager.SetUpdateForTask(
		state.ExecutionTaskStatus.TaskKey,
		state,
	)
	runtimeinfo.LogInfo("(RESP. TO: -> GITHUB-COLLECTOR) COMPLETED OK")
	context.AbortWithStatus(http.StatusOK)
}

func (service *CollectorService) restHandlerUpdateRepositoryIssues(context *gin.Context) {
	state := new(jsonSendFromCollectorRepositoryIssues)
	if err := context.BindJSON(state); err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		context.AbortWithStatus(http.StatusLocked)
		return
	}
	service.taskManager.SetUpdateForTask(
		state.ExecutionTaskStatus.TaskKey,
		state,
	)
	runtimeinfo.LogInfo("(RESP. TO: -> GITHUB-COLLECTOR) COMPLETED OK")
	context.AbortWithStatus(http.StatusOK)
}