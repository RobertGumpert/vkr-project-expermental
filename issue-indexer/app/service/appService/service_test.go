package appService

import (
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"issue-indexer/app/config"
	"testing"
)

func initDependents() (*config.Config, repository.IRepository) {
	CONFIG := config.NewConfig().ReadWithPath("C:/VKR/vkr-project-expermental/issue-indexer/data/config/config.json")
	POSTGRES := repository.NewSQLRepository(
		repository.SQLCreateConnection(
			repository.TypeStoragePostgres,
			repository.DSNPostgres,
			nil,
			CONFIG.Postgres.Username,
			CONFIG.Postgres.Password,
			CONFIG.Postgres.DbName,
			CONFIG.Postgres.Port,
			CONFIG.Postgres.Ssl,
		),
	)
	return CONFIG, POSTGRES
}

// 30 - 3,7,5,60,49,26,20,45
// 33 - 49,48,41,56,58,34
// 144
func TestIndexingGroupRepositoriesFlow(t *testing.T) {
	fakeChannel := make(chan bool)
	c, p := initDependents()
	service := NewAppService(
		p,
		c,
	)
	err := service.CreateTaskCompareGroup(&jsonSendFromGateCompareGroup{
		"test_0",
		144,
		[]uint{137, 126},
	})
	if err != nil {
		runtimeinfo.LogError(err)
	}
	err = service.CreateTaskCompareGroup(&jsonSendFromGateCompareGroup{
		"test_1",
		33,
		[]uint{49,48,41,56,58,34},
	})
	if err != nil {
		runtimeinfo.LogError(err)
	}
	err = service.CreateTaskCompareGroup(&jsonSendFromGateCompareGroup{
		"test_2",
		33,
		[]uint{3,7,5,60,49,26,20,45},
	})
	if err != nil {
		runtimeinfo.LogError(err)
	}
	fakeChannel <- true
}
