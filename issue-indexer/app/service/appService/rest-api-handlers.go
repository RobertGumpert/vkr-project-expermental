package appService

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (service *AppService) ConcatTheirRestHandlers(engine *gin.Engine) {
	api := engine.Group("/api/task")
	api.GET(
		"/get/state",
		service.restHandlerGateState,
	)
	api.POST(
		"/compare/all/beside",
		service.restHandlerCompareBeside,
	)
	api.POST(
		"/compare/group",
		service.restHandlerCompareGroup,
	)
}

func (service *AppService) restHandlerGateState(ctx *gin.Context) {
	queueIsFilled := service.QueueIsFilled()
	if queueIsFilled {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	ctx.AbortWithStatus(http.StatusOK)
	return
}

func (service *AppService) restHandlerCompareBeside(ctx *gin.Context) {
	state := new(jsonSendFromCompareBeside)
	if err := ctx.BindJSON(state); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.CreateTaskCompareBeside(state)
	if err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	ctx.AbortWithStatus(http.StatusOK)
	return
}

func (service *AppService) restHandlerCompareGroup(ctx *gin.Context) {
	state := new(jsonSendFromGateCompareGroup)
	if err := ctx.BindJSON(state); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.CreateTaskCompareGroup(state)
	if err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	ctx.AbortWithStatus(http.StatusOK)
	return
}
