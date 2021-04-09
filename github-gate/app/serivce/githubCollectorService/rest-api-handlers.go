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
	updateTaskStateHandlers.POST(
		"/repositories/by/keyword",
		service.restHandlerUpdateRepositoriesByKeyWord,
	)
}

func (service *CollectorService) restHandlerUpdateDescriptionsRepositories(context *gin.Context) {
	state := new(jsonSendFromCollectorDescriptionsRepositories)
	if err := context.BindJSON(state); err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		context.AbortWithStatus(http.StatusLocked)
		return
	}
	runtimeinfo.LogInfo("COLLECTOR -> : UPDATE TASK [",state.ExecutionTaskStatus.TaskKey,"]")
	service.taskManager.SetUpdateForTask(
		state.ExecutionTaskStatus.TaskKey,
		state,
	)
	context.AbortWithStatus(http.StatusOK)
}

func (service *CollectorService) restHandlerUpdateRepositoryIssues(context *gin.Context) {
	state := new(jsonSendFromCollectorRepositoryIssues)
	if err := context.BindJSON(state); err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		context.AbortWithStatus(http.StatusLocked)
		return
	}
	runtimeinfo.LogInfo("COLLECTOR -> : UPDATE TASK [",state.ExecutionTaskStatus.TaskKey,"]")
	service.taskManager.SetUpdateForTask(
		state.ExecutionTaskStatus.TaskKey,
		state,
	)
	context.AbortWithStatus(http.StatusOK)
}

func (service *CollectorService) restHandlerUpdateRepositoriesByKeyWord(context *gin.Context) {
	state := new(jsonSendFromCollectorRepositoriesByKeyWord)
	if err := context.BindJSON(state); err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		context.AbortWithStatus(http.StatusLocked)
		return
	}
	runtimeinfo.LogInfo("COLLECTOR -> : UPDATE TASK [",state.ExecutionTaskStatus.TaskKey,"]")
	service.taskManager.SetUpdateForTask(
		state.ExecutionTaskStatus.TaskKey,
		state,
	)
	context.AbortWithStatus(http.StatusOK)
}