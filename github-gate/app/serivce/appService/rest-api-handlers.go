package appService

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (service *AppService) ConcatTheirRestHandlers(engine *gin.Engine) {
	apiGroup := engine.Group("/api")
	{
		downloadGroup := apiGroup.Group("/download/and/analyze")
		{
			downloadGroup.POST("/new/repository/exist/keyword", service.restHandlerNewRepositoryExistKeyword)
			downloadGroup.POST("/new/repository/new/keyword", service.restHandlerNewRepositoryNewKeyword)
		}
		reAnalyzeGroup := apiGroup.Group("/reanalyze")
		{
			reAnalyzeGroup.POST("/exist/repository", service.restHandlerReanalyzeExistRepository)
		}
	}
}

func (service *AppService) restHandlerNewRepositoryExistKeyword(ctx *gin.Context) {
	state := new(JsonNewRepositoryWithExistKeyword)
	if err := ctx.BindJSON(state); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.DownloadAndAnalyzeNewRepositoryWithExistKeyword(
		state,
	)
	if err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
	}
	ctx.AbortWithStatus(http.StatusOK)
}

func (service *AppService) restHandlerNewRepositoryNewKeyword(ctx *gin.Context) {
	state := new(JsonNewRepositoryWithNewKeyword)
	if err := ctx.BindJSON(state); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.DownloadAndAnalyzeNewRepositoryWithNewKeyword(
		state,
	)
	if err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
	}
	ctx.AbortWithStatus(http.StatusOK)
}

func (service *AppService) restHandlerReanalyzeExistRepository(ctx *gin.Context) {
	state := new(JsonExistRepository)
	if err := ctx.BindJSON(state); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.ReanalyzeExistRepository(
		state,
	)
	if err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
	}
	ctx.AbortWithStatus(http.StatusOK)
}
