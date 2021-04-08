package appService

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (service *AppService) ConcatTheirRestHandlers(engine *gin.Engine) {
	apiGroup := engine.Group("/api")
	{
		collectorGroup := apiGroup.Group("/download/repositories")
		{
			collectorGroup.POST("/by/name", service.restHandlerDownloadRepositoriesByName)
			collectorGroup.POST("/by/keyword", service.restHandlerDownloadRepositoriesByKeyWord)
		}
	}
}

func (service *AppService) restHandlerDownloadRepositoriesByName(ctx *gin.Context) {
	requestData := new(JsonSingleTaskDownloadRepositoriesByName)
	if err := ctx.BindJSON(requestData); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.CreateTaskDownloadRepositoriesByNames(
		requestData,
	)
	if err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
	}
	ctx.AbortWithStatus(http.StatusOK)
}

func (service *AppService) restHandlerDownloadRepositoriesByKeyWord(ctx *gin.Context) {
	requestData := new(JsonSingleTaskDownloadRepositoriesByKeyWord)
	if err := ctx.BindJSON(requestData); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.CreateTaskDownloadRepositoriesByKeyWord(
		requestData,
	)
	if err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
	}
	ctx.AbortWithStatus(http.StatusOK)
}