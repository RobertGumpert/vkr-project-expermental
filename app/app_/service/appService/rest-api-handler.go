package appService

import (
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (service *AppService) ConcatTheirRestHandlers(root string, engine *gin.Engine) {
	engine.Use(
		cors.Default(),
	)
	engine.Static("../js", root+"/data/assets/js")
	engine.Static("../css", root+"/data/assets/css")
	engine.Static("../images/", root+"/data/assets/images")
	engine.LoadHTMLGlob(root + "/data/assets/html/*.html")
	//
	engine.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index.html", nil)
		return
	})
	//
	taskApi := engine.Group("/task/api/update")
	{
		taskApi.POST("/nearest/repositories", service.restHandlerUpdateTaskStateNearestRepositories)
	}
	userEndpoints := engine.Group("/get")
	{
		userEndpoints.GET("/:digest", service.restHandlerDigest)
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
	state := new(JsonCreateTaskFindNearestRepositories)
	if err := ctx.BindJSON(state); err != nil {
		runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	jsonBody, err := service.FindNearestRepositories(state)
	if err != nil {
		if err == ErrorRequestReceivedLater {
			jsonBody := &JsonResultTaskFindNearestRepositories{TaskState: &JsonStateTask{
				IsDefer: true,
			}}
			ctx.AbortWithStatusJSON(http.StatusOK, jsonBody)
			return
		} else {
			ctx.AbortWithStatus(http.StatusLocked)
			return
		}
	}
	hash, err := jsonBody.encodeHash()
	if err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	url := strings.Join([]string{
		"get",
		hash,
	}, "/")
	jsonBody.TaskState = &JsonStateTask{
		IsDefer:  false,
		Endpoint: url,
	}
	ctx.AbortWithStatusJSON(http.StatusOK, jsonBody)
	return
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
	hash := ctx.Param("digest")
	state := new(JsonResultTaskFindNearestRepositories)
	err := state.decodeHash(hash)
	if err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}

	ctx.HTML(
		http.StatusOK,
		"nearest-repositories-template.html",
		state,
	)
	return
}
