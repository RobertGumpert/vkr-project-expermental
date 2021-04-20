package main

import (
	"app/app_/config"
	"app/app_/service/appService"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"runtime"
)

var (
	POSTGRES   *repository.SQLRepository
	CONFIG     *config.Config
	SERVER     *server
	APPSERVICE *appService.AppService
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

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
	defer func() {
		if err := POSTGRES.CloseConnection(); err != nil {
			runtimeinfo.LogFatal(err)
		}
	}()
	SERVER = NewServer(CONFIG)
	APPSERVICE = appService.NewAppService(
		POSTGRES,
		CONFIG,
		SERVER.engine,
	)
	SERVER.RunServer()
}
