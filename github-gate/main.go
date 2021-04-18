package main

import (
	"github-gate/app/config"
	"github-gate/app/serivce/appService"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
)

var (
	APPSERVICE *appService.AppService
	POSTGRES   *repository.SQLRepository
	CONFIG     *config.Config
	SERVER     *server
)

func main() {
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
