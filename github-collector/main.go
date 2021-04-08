package main

import (
	"github-collector/app/config"
	"github-collector/app/service/appService"
)

var (
	CONFIG *config.Config
	SERVER *server
	APP    *appService.AppService
)

func main() {
	CONFIG = config.NewConfig().Read()
	APP = appService.NewAppService(CONFIG)
	SERVER = NewServer(CONFIG)
	APP.ConcatTheirRestHandlers(SERVER.engine)
	SERVER.RunServer()
}
