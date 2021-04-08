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
			collectorGroup.POST("/by/name", service.apiDownloadRepositoriesByName)
			collectorGroup.POST("/by/keyword")
		}
	}
}

func (service *AppService) apiDownloadRepositoriesByName(ctx *gin.Context) {
	requestData := new(ApiJsonDownloadRepositoriesByName)
	if err := ctx.BindJSON(requestData); err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := service.CreateApiTaskDownloadRepositoriesByNames(
		requestData,
	)
	if err != nil {
		ctx.AbortWithStatus(http.StatusLocked)
	}
	ctx.AbortWithStatus(http.StatusOK)
}
