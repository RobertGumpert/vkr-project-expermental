package appService

import (
	"github-collector/pckg/runtimeinfo"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (service *AppService) ConcatTheirRestHandlers(engine *gin.Engine) {
	handlers := engine.Group("/api/task")
	handlers.GET(
		"/get/state",
		service.restHandlerGetState,
	)
	handlers.POST(
		"/repositories/descriptions",
		service.restHandlerCreateTaskRepositoriesDescriptions,
	)
	handlers.POST(
		"/repository/issues",
		service.restHandlerCreateTaskRepositoryIssues,
	)
	handlers.POST(
		"/repositories/by/keyword",
		service.restHandlerCreateTaskRepositoriesByKeyWord,
	)
}

func (service *AppService) restHandlerGetState(ctx *gin.Context) {
	if err, all := service.GithubClient.GetState(); err != nil {
		runtimeinfo.LogInfo("count all task : [", all, "];")
		ctx.AbortWithStatus(http.StatusLocked)
		return
	} else {
		runtimeinfo.LogInfo("count all task : [", all, "];")
		ctx.AbortWithStatus(http.StatusOK)
		return
	}
}

func (service *AppService) restHandlerCreateTaskRepositoriesDescriptions(ctx *gin.Context) {
	state := new(JsonCreateTaskRepositoriesDescriptions)
	if err := ctx.BindJSON(state); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	if err := service.CreateTaskRepositoriesDescriptions(state); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	ctx.AbortWithStatus(http.StatusOK)
}

func (service *AppService) restHandlerCreateTaskRepositoryIssues(ctx *gin.Context) {
	state := new(JsonCreateTaskRepositoryIssues)
	if err := ctx.BindJSON(state); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	if err := service.CreateTaskRepositoryIssues(state); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	ctx.AbortWithStatus(http.StatusOK)
}

func (service *AppService) restHandlerCreateTaskRepositoriesByKeyWord(ctx *gin.Context) {
	state := new(JsonCreateTaskRepositoriesByKeyWord)
	if err := ctx.BindJSON(state); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	if err := service.CreateTaskRepositoriesByKeyWord(state); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	ctx.AbortWithStatus(http.StatusOK)
}