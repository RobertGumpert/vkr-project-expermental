package main

import (
	"issue-indexer/app/config"
	"issue-indexer/app/scanner"
	"time"
)

var (
	SCANNER *scanner.DatabaseUpdates
	SERVER  *server
	CONFIG  *config.Config
)

func main() {
	SCANNER = scanner.NewDatabaseUpdates(2 * time.Second)
	SCANNER.Run()
	CONFIG = config.NewConfig().Read()
	SERVER = NewServer(CONFIG)
	SERVER.RunServer()
}
