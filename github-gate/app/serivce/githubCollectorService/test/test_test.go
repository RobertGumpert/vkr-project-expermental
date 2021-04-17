package test

import (
	"github-gate/app/config"
	"github-gate/app/serivce/githubCollectorService"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/gin-gonic/gin"
	"testing"
)

var storageProvider = repository.SQLCreateConnection(
	repository.TypeStoragePostgres,
	repository.DSNPostgres,
	nil,
	"postgres",
	"toster123",
	"vkr-db",
	"5432",
	"disable",
)

func connect() repository.IRepository {
	sqlRepository := repository.NewSQLRepository(
		storageProvider,
	)
	return sqlRepository
}


func TestMigration(t *testing.T) {
	storageProvider.SqlDB.Exec("drop table repository_models cascade")
	storageProvider.SqlDB.Exec("drop table issue_models cascade")
	storageProvider.SqlDB.Exec("drop table nearest_issues_models cascade")
	storageProvider.SqlDB.Exec("drop table nearest_repositories_models cascade")
	storageProvider.SqlDB.Exec("drop table repositories_key_words_models cascade")
	storageProvider.SqlDB.Exec("drop table number_issue_intersections_models cascade")
	_ = connect()
}

func createFakeTaskService(c *config.Config) (*githubCollectorService.CollectorService, *gin.Engine) {
	db := connect()
	server := gin.Default()
	service := githubCollectorService.NewCollectorService(
		db,
		c,
		server,
	)
	return service, server
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
