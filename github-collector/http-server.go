package main

import (
	"github-collector/pckg/runtimeinfo"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

type server struct {
	config     *config
	engine     *gin.Engine
	appService *appService
	RunServer  func()
}

func NewServer(config *config, app *appService) *server {
	s := &server{
		config: config,
	}
	//
	engine, run := s.createServerEngine(s.config.Port)
	s.RunServer = run
	s.engine = engine
	s.appService = app
	//
	s.engine.GET("/get/state", s.getState)
	s.engine.POST("/get/repos/by/url", s.getRepositoriesByURL)
	s.engine.POST("/get/repos/issues", s.getRepositoryIssue)
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
			log.Fatal(err)
		}
	}
}

//
//----------------------------------------------HANDLERS----------------------------------------------------------------
//

func (s *server) getState(ctx *gin.Context) {
	if err, all := s.appService.GITHUBClient.GetState(); err != nil {
		runtimeinfo.LogInfo("count all task : [", all, "];")
		ctx.AbortWithStatus(http.StatusLocked)
		return
	} else {
		runtimeinfo.LogInfo("count all task : [", all, "];")
		ctx.AbortWithStatus(http.StatusOK)
		return
	}
}

func (s *server) getRepositoriesByURL(ctx *gin.Context) {
	data := new(CreateTaskReposByURL)
	if err := ctx.BindJSON(data); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err := s.appService.GetReposByURLS(data.TaskKey, data.URLS)
	if err != nil {
		// 423
		ctx.AbortWithStatus(http.StatusLocked)
		return
	} else {
		// 200
		ctx.AbortWithStatus(http.StatusOK)
		return
	}
}

func (s *server) getRepositoryIssue(ctx *gin.Context) {
	data := new(CreateTaskRepositoryIssues)
	if err := ctx.BindJSON(data); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err := s.appService.GetRepositoryIssues(data.TaskKey, data.URL)
	if err != nil {
		// 423
		ctx.AbortWithStatus(http.StatusLocked)
		return
	} else {
		// 200
		ctx.AbortWithStatus(http.StatusOK)
		return
	}
}
