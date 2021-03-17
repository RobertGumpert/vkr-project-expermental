package githubTasksService

import (
	"github-gate/app/models/interapplicationModels/githubCollectorModels"
	"github-gate/pckg/runtimeinfo"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (service *GithubTasksService) RestRepositoriesDescriptionByURL(context *gin.Context) {
	state := new(githubCollectorModels.UpdateTaskRepositoriesByURLS)
	if err := context.BindJSON(state); err != nil {
		runtimeinfo.LogInfo("(FROM: GITHUB-COLLECTOR->) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		context.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.updateTaskRepositoriesDescriptionByURL(state)
	if err != nil {
		runtimeinfo.LogInfo("UPDATE TASK [", state.ExecutionTaskStatus.TaskKey, "] COMPLETED WITH ERROR: ", err)
		context.AbortWithStatus(http.StatusLocked)
		return
	}
	context.AbortWithStatus(http.StatusOK)
}
