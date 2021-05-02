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
	engine.Static("/js", root+"/data/assets/js")
	engine.Static("/css", root+"/data/assets/css")
	engine.Static("/images", root+"/data/assets/images")
	engine.LoadHTMLGlob(root + "/data/assets/*.html")
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
	gettingEndpoints := engine.Group("/get")
	{
		gettingEndpoints.GET("/:digest", service.restHandlerNearestRepositoriesDigest)
		gettingEndpoints.POST("/nearest/repositories", service.restHandlerGetNearestRepositories)
		gettingEndpoints.GET("/nearest/issues/:userRepository/with/:nearestRepository", service.restHandlerGetNearestIssues)
	}
	notificationEndpoints := engine.Group("/notification")
	{
		notificationEndpoints.GET("/:digest", service.restHandlerNotificationDigest)
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
			hash, err := state.encodeHash()
			if err != nil {
				ctx.AbortWithStatus(http.StatusLocked)
				return
			}
			jsonBody := &JsonResultTaskFindNearestRepositories{
				TaskState: &JsonStateTask{
					IsDefer:  true,
					Endpoint: strings.Join([]string{"/notification", hash}, "/"),
				},
			}
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
	userRepository := ctx.Param("userRepository")
	nearestRepository := ctx.Param("nearestRepository")
	jsonModel, err := service.GetNearestIssuesInPairNearestRepositories(
		userRepository,
		nearestRepository,
	)
	if err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	ctx.HTML(
		http.StatusOK,
		"nearest-issues-template.html",
		jsonModel,
	)
	return
}

func (service *AppService) restHandlerNearestRepositoriesDigest(ctx *gin.Context) {
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

func (service *AppService) restHandlerNotificationDigest(ctx *gin.Context) {
	hash := ctx.Param("digest")
	state := new(JsonCreateTaskFindNearestRepositories)
	err := state.decodeHash(hash)
	if err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	ctx.HTML(
		http.StatusOK,
		"defer-result-message-template.html",
		state,
	)
	return
}
