package main

var(
	CONFIG *config
	SERVER *server
	APP *appService
)


func main() {
	CONFIG = NewConfig().Read()
	APP = NewAppService(CONFIG)
	SERVER = NewServer(CONFIG, APP)
	SERVER.RunServer()
}
