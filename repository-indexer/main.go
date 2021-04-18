package main

import (
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"os"
	"repository-indexer/app/config"
	"repository-indexer/app/service/appService"
	"repository-indexer/app/service/hashRepository"
	"runtime"
)

var (
	LOCALDATABASE repository.IRepository
	POSTGRES      *repository.SQLRepository
	CONFIG        *config.Config
	SERVER        *server
	APPSERVICE    *appService.AppService
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//
	var err error
	runtimeinfo.LogInfo("START.")
	CONFIG = config.NewConfig().Read()
	POSTGRES = repository.NewSQLRepository(
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
	LOCALDATABASE, err = hashRepository.NewLocalHashStorage(getRootPathProject())
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	defer func() {
		if err := POSTGRES.CloseConnection(); err != nil {
			runtimeinfo.LogFatal(err)
		}
		if err := LOCALDATABASE.CloseConnection(); err != nil {
			runtimeinfo.LogFatal(err)
		}
	}()

	APPSERVICE, err = appService.NewAppService(
		CONFIG,
		POSTGRES,
		LOCALDATABASE,
	)
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	SERVER = NewServer(CONFIG)
	APPSERVICE.ConcatTheirRestHandlers(SERVER.engine)
	SERVER.RunServer()
}

func getRootPathProject() (path string) {
	directory, err := os.Getwd()
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	return directory
}
