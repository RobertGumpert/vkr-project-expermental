package main

import (
	"github-gate/app/config"
	"github-gate/app/serivce"
)

var(
	CONFIG *config.Config
	SERVER *server
	APP *serivce.AppService
)


func main() {
	CONFIG = config.NewConfig().Read()
	APP = serivce.NewAppService(CONFIG)
	SERVER = NewServer(CONFIG, APP)
	SERVER.RunServer()
}