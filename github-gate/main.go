package main

import (
	"github-gate/app/config"
	"github-gate/app/repository"
	"github-gate/app/serivce"
	"github-gate/pckg/runtimeinfo"
)

var (
	POSTGRES *repository.SQLRepository
	CONFIG   *config.Config
	SERVER   *server
	APP      *serivce.AppService
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
	APP = serivce.NewAppService(CONFIG)
	SERVER = NewServer(CONFIG, APP)
	SERVER.RunServer()
}
