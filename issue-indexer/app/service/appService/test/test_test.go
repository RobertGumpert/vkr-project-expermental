package test

import (
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/gin-gonic/gin"
	"issue-indexer/app/config"
	"issue-indexer/app/service/appService"
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

func TestTruncate(t *testing.T) {
	_ = connect()
	storageProvider.SqlDB.Exec("TRUNCATE TABLE repositories CASCADE")
	storageProvider.SqlDB.Exec("TRUNCATE TABLE issues CASCADE")
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

func createFakeConfig() *config.Config {
	return &config.Config{
		Port:                                  "44000",
		MaxCountRunnableTasks:                 100,
		MaxCountThreads:                       100,
		MinimumTextCompletenessThreshold:      70.0,
		MaximumDurationDatabaseQueryInMinutes: 1,
		GithubGateAddress:                     "http://127.0.0.1:54000",
		GithubGateEndpoints: struct {
			SendResultTaskCompareGroup  string `json:"send_result_task_compare_group"`
			SendResultTaskCompareBeside string `json:"send_result_task_compare_beside"`
		}{
			SendResultTaskCompareGroup:  "/task/api/issueindexer/update/compare/group",
			SendResultTaskCompareBeside: "/task/api/issueindexer/update/compare/beside",
		},
	}
}

func createFakeService(c *config.Config) (*appService.AppService, *gin.Engine) {
	db := connect()
	server := gin.Default()
	service := appService.NewAppService(
		db,
		c,
	)
	service.ConcatTheirRestHandlers(server)
	return service, server
}
