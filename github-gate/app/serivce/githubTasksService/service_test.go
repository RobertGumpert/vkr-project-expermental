package githubTasksService

import (
	"github-gate/app/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"
)

func createFakeHttpServer(service *GithubTasksService) *gin.Engine {
	router := gin.Default()

	router.POST("api/collector/task/create/repos/by/url", service.RestCreateTaskRepositoriesByURL)
	router.POST("api/collector/task/create/issue/by/repo", service.RestCreateRepositoryIssues)

	router.POST("api/collector/task/result/repos/by/url", service.RestUpdateRepositoriesDescriptionByURL)
	router.POST("api/collector/task/result/issue/by/repo", service.RestUpdateRepositoryIssues)

	return router
}

func createFakeTaskService(c *config.Config) *GithubTasksService {
	service := NewGithubTasksService(
		c,
		new(http.Client),
	)
	return service
}

func createFakeConfig() *config.Config {
	return &config.Config{
		Port:                              "54000",
		SizeQueueTasksForGithubCollectors: 10000,
		GithubCollectorsAddresses: []string{
			"http://127.0.0.1:54100",
		},
	}
}

func TestTaskFlow(t *testing.T) {
	c := createFakeConfig()
	service := createFakeTaskService(c)
	server := createFakeHttpServer(service)
	err := server.Run(":" + c.Port)
	if err != nil {
		t.Fatal(err)
	}
}