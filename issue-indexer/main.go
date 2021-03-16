package main

import (
	"issue-indexer/app/config"
	"issue-indexer/app/repository"
	"issue-indexer/app/service/tasksService"
	"issue-indexer/pckg/runtimeinfo"
)

var (
	POSTGRES    *repository.SQLRepository
	SERVER      *server
	CONFIG      *config.Config
	TASKSERVICE *tasksService.TasksService
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
	TASKSERVICE = tasksService.NewTasksService(CONFIG, POSTGRES)
	SERVER = NewServer(CONFIG, TASKSERVICE)
	SERVER.RunServer()
}
