package githubTasksService

import (
	"github-gate/app/models/interapplicationModels/githubCollectorModels"
	"github-gate/pckg/runtimeinfo"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (service *GithubTasksService) RestUpdateRepositoriesDescriptionByURL(context *gin.Context) {
	state := new(githubCollectorModels.UpdateTaskRepositoriesByURLS)
	if err := context.BindJSON(state); err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		context.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.updateTaskRepositoriesDescriptionByURL(state)
	if err != nil {
		runtimeinfo.LogError("(RESP. TO: ->GITHUB-COLLECTOR) TASK [", state.ExecutionTaskStatus.TaskKey, "] SEND RESPONSE: 423")
		context.AbortWithStatus(http.StatusLocked)
		return
	}
	runtimeinfo.LogInfo("(RESP. TO: ->GITHUB-COLLECTOR) TASK [", state.ExecutionTaskStatus.TaskKey, "] SEND RESPONSE: 200")
	context.AbortWithStatus(http.StatusOK)
}

func (service *GithubTasksService) RestUpdateRepositoryIssues(context *gin.Context) {
	state := new(githubCollectorModels.UpdateTaskRepositoryIssues)
	if err := context.BindJSON(state); err != nil {
		runtimeinfo.LogError("(RESP. TO: ->GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		context.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.updateTaskRepositoryIssues(state)
	if err != nil {
		runtimeinfo.LogError("(RESP. TO: ->GITHUB-COLLECTOR) TASK [", state.ExecutionTaskStatus.TaskKey, "] SEND RESPONSE: 423")
		context.AbortWithStatus(http.StatusLocked)
		return
	}
	runtimeinfo.LogInfo("(RESP. TO: ->GITHUB-COLLECTOR) TASK [", state.ExecutionTaskStatus.TaskKey, "] SEND RESPONSE: 200")
	context.AbortWithStatus(http.StatusOK)
}

func (service *GithubTasksService) RestCreateTaskRepositoriesByURL(ctx *gin.Context) {
	task := new(githubCollectorModels.CreateTaskRepositoriesByURLS)
	if err := ctx.BindJSON(task); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	queueIsBusy, sendTaskToGithubCollector := service.CreateTaskRepositoriesDescriptions(nil, task.Repositories)
	if !queueIsBusy() {
		ctx.AbortWithStatus(http.StatusLocked)
	}
	sendTaskToGithubCollector()
	ctx.AbortWithStatus(http.StatusOK)
	return
}

func (service *GithubTasksService) RestCreateRepositoryIssues(ctx *gin.Context) {
	task := new(githubCollectorModels.CreateTaskRepositoriesByURLS)
	if err := ctx.BindJSON(task); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	queueIsBusy, sendTaskToGithubCollector := service.CreateTaskRepositoriesAndTheirIssues(nil, task.Repositories)
	if !queueIsBusy() {
		ctx.AbortWithStatus(http.StatusLocked)
	}
	sendTaskToGithubCollector()
	ctx.AbortWithStatus(http.StatusOK)
	return
}
