package appService

import (
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (service *AppService) ConcatTheirRestHandlers(engine *gin.Engine) {
	taskApi := engine.Group("/task/api/update")
	{
		taskApi.POST("/nearest/repositories", service.restHandlerUpdateTaskStateNearestRepositories)
	}
	userEndpoints := engine.Group("/get")
	{
		userEndpoints.POST("/:digest", service.restHandlerDigest)
		userEndpoints.POST("/nearest/repositories", service.restHandlerGetNearestRepositories)
		userEndpoints.POST("/nearest/issues", service.restHandlerGetNearestIssues)
	}
}

func (service *AppService) restHandlerUpdateTaskStateNearestRepositories(ctx *gin.Context) {
	state := new(JsonFromGetNearestRepositories)
	if err := ctx.BindJSON(state); err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	service.SendDeferResponseToClient(state)
	ctx.AbortWithStatus(http.StatusOK)
}

func (service *AppService) restHandlerGetNearestRepositories(ctx *gin.Context) {
	//state := new(JsonFromGetNearestRepositories)
	//if err := ctx.BindJSON(state); err != nil {
	//	runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
	//	ctx.AbortWithStatus(http.StatusLocked)
	//	return
	//}
	//service.SendDeferResponseToClient(state)
	//ctx.AbortWithStatus(http.StatusOK)
}

func (service *AppService) restHandlerGetNearestIssues(ctx *gin.Context) {
	//state := new(JsonFromGetNearestRepositories)
	//if err := ctx.BindJSON(state); err != nil {
	//	runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
	//	ctx.AbortWithStatus(http.StatusLocked)
	//	return
	//}
	//service.SendDeferResponseToClient(state)
	//ctx.AbortWithStatus(http.StatusOK)
}


func (service *AppService) restHandlerDigest(ctx *gin.Context) {
	//state := new(JsonFromGetNearestRepositories)
	//if err := ctx.BindJSON(state); err != nil {
	//	runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
	//	ctx.AbortWithStatus(http.StatusLocked)
	//	return
	//}
	//service.SendDeferResponseToClient(state)
	//ctx.AbortWithStatus(http.StatusOK)
}