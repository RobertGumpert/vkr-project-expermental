package appService

import (
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"log"
	"repository-indexer/app/config"
	"repository-indexer/app/service/hashRepository"
	"strconv"
	"testing"
)

func initDependents() (*config.Config, repository.IRepository, repository.IRepository) {
	CONFIG := config.NewConfig().ReadWithPath("C:/VKR/vkr-project-expermental/repository-indexer/data/config/config.json")
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
	LOCALDATABASE, err := hashRepository.NewLocalHashStorage("C:/VKR/vkr-project-expermental/repository-indexer")
	if err != nil {
		log.Fatal(err)
	}
	return CONFIG, POSTGRES, LOCALDATABASE
}

func TestReindexingForAllFlow(t *testing.T) {
	fakeChannel := make(chan bool)
	c, p, l := initDependents()
	service, err := NewAppService(c, p, l)
	if err != nil {
		t.Fatal(err)
	}
	for i := int64(0); i < c.MaximumSizeOfQueue+1; i++ {
		err = service.AddTask(&jsonSendFromGateReindexingForAll{
			TaskKey: "test_" + strconv.Itoa(int(i)),
		}, taskTypeReindexingForAll)
		if err != nil {
			runtimeinfo.LogError(err)
		}
	}
	fakeChannel <- true
}

func TestReindexingForRepositoryFlow(t *testing.T) {
	fakeChannel := make(chan bool)
	c, p, l := initDependents()
	service, err := NewAppService(c, p, l)
	if err != nil {
		t.Fatal(err)
	}
	for i := int64(0); i < c.MaximumSizeOfQueue+1; i++ {
		err = service.AddTask(&jsonSendFromGateReindexingForRepository{
			TaskKey:      "test_" + strconv.Itoa(int(i)),
			RepositoryID: 51,
		}, taskTypeReindexingForRepository)
		if err != nil {
			runtimeinfo.LogError(err)
		}
	}
	fakeChannel <- true
}

func TestReindexingForGroupRepositoriesFlow(t *testing.T) {
	fakeChannel := make(chan bool)
	c, p, l := initDependents()
	service, err := NewAppService(c, p, l)
	if err != nil {
		t.Fatal(err)
	}
	for i := int64(0); i < c.MaximumSizeOfQueue+1; i++ {
		err = service.AddTask(&jsonSendFromGateReindexingForGroupRepositories{
			TaskKey:      "test_" + strconv.Itoa(int(i)),
			RepositoryID: []uint{51, 17},
		}, taskTypeReindexingForGroupRepositories)
		if err != nil {
			runtimeinfo.LogError(err)
		}
	}
	fakeChannel <- true
}

func TestGetNearestRepositoriesFlow(t *testing.T) {
	fakeChannel := make(chan bool)
	service, err := NewAppService(initDependents())
	if err != nil {
		t.Fatal(err)
	}
	output := service.RepositoryNearest(&jsonInputNearestRepositoriesForRepository{
		RepositoryID: 51,
	})
	runtimeinfo.LogInfo(*output)
	fakeChannel <- true
}

func TestGetKeywordFlow(t *testing.T) {
	fakeChannel := make(chan bool)
	service, err := NewAppService(initDependents())
	if err != nil {
		t.Fatal(err)
	}
	output := service.WordIsExist(&jsonInputWordIsExist{
		Word: "vpn",
	})
	runtimeinfo.LogInfo(*output)
	fakeChannel <- true
}
