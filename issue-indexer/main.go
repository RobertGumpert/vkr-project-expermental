package main

import (
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"issue-indexer/app/config"
	"issue-indexer/app/service/appService"
	"runtime"
)

var (
	APPSERVICE *appService.AppService
	POSTGRES   *repository.SQLRepository
	SERVER     *server
	CONFIG     *config.Config
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
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
	defer func() {
		if err := POSTGRES.CloseConnection(); err != nil {
			runtimeinfo.LogFatal(err)
		}
	}()
	APPSERVICE = appService.NewAppService(POSTGRES, CONFIG)
	SERVER = NewServer(CONFIG)
	APPSERVICE.ConcatTheirRestHandlers(SERVER.engine)
	SERVER.RunServer()
}
