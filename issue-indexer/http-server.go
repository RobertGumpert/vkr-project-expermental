package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"issue-indexer/app/config"
	"issue-indexer/pckg/runtimeinfo"
	"strings"
)

type server struct {
	config    *config.Config
	engine    *gin.Engine
	RunServer func()
}

func NewServer(config *config.Config) *server {
	s := &server{
		config: config,
	}
	//
	engine, run := s.createServerEngine(s.config.Port)
	s.RunServer = run
	s.engine = engine
	//
	return s
}

func (s *server) createServerEngine(port ...string) (*gin.Engine, func()) {
	var serverPort = ""
	if len(port) != 0 {
		if !strings.Contains(port[0], ":") {
			serverPort = strings.Join([]string{
				":",
				port[0],
			}, "")
		}
	}
	engine := gin.Default()
	engine.Use(
		cors.Default(),
	)
	return engine, func() {
		var err error
		if serverPort != "" {
			err = engine.Run(serverPort)
		} else {
			err = engine.Run()
		}
		if err != nil {
			runtimeinfo.LogFatal(err)
		}
	}
}
