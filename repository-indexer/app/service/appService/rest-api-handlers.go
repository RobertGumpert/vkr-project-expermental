package appService

import (
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (service *AppService) ConcatTheirRestHandlers(engine *gin.Engine) {
	taskApi := engine.Group("/api/task")
	taskApi.GET(
		"/get/state",
		service.restHandlerGetState,
	)
	taskApi.POST(
		"/reindexing/for/all",
		service.restHandlerReindexingForAll,
	)
	taskApi.POST(
		"/reindexing/for/repository",
		service.restHandlerReindexingForRepository,
	)
	taskApi.POST(
		"/reindexing/for/group/repositories",
		service.restHandlerReindexingForGroupRepositories,
	)
	api := engine.Group("/api/get")
	api.POST(
		"/word",
		func(ctx *gin.Context) {
			state := new(jsonInputWordIsExist)
			if err := ctx.BindJSON(state); err != nil {
				runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
				ctx.AbortWithStatus(http.StatusLocked)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusOK, service.WordIsExist(state))
			return
		},
	)
	api.POST(
		"/nearest/for/repository",
		func(ctx *gin.Context) {
			state := new(jsonInputNearestRepositoriesForRepository)
			if err := ctx.BindJSON(state); err != nil {
				runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
				ctx.AbortWithStatus(http.StatusLocked)
				return
			}
			ctx.AbortWithStatusJSON(http.StatusOK, service.RepositoryNearest(state))
			return
		},
	)
}

func (service *AppService) restHandlerGetState(ctx *gin.Context) {
	if service.QueueIsFilled() {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	ctx.AbortWithStatus(http.StatusOK)
}

func (service *AppService) restHandlerReindexingForAll(ctx *gin.Context) {
	state := new(jsonSendFromGateReindexingForAll)
	if err := ctx.BindJSON(state); err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.AddTask(state, taskTypeReindexingForAll)
	if err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	ctx.AbortWithStatus(http.StatusOK)
}

func (service *AppService) restHandlerReindexingForRepository(ctx *gin.Context) {
	state := new(jsonSendFromGateReindexingForRepository)
	if err := ctx.BindJSON(state); err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.AddTask(state, taskTypeReindexingForRepository)
	if err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	ctx.AbortWithStatus(http.StatusOK)
}

func (service *AppService) restHandlerReindexingForGroupRepositories(ctx *gin.Context) {
	state := new(jsonSendFromGateReindexingForGroupRepositories)
	if err := ctx.BindJSON(state); err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.AddTask(state, taskTypeReindexingForGroupRepositories)
	if err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	ctx.AbortWithStatus(http.StatusOK)
}
