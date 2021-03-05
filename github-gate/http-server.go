package main

import (
	"github-gate/app/config"
	"github-gate/app/models/createTaskModel"
	"github-gate/app/models/updateTaskModel"
	"github-gate/app/serivce"
	"github-gate/pckg/runtimeinfo"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type server struct {
	config     *config.Config
	engine     *gin.Engine
	appService *serivce.AppService
	RunServer  func()
}

func NewServer(config *config.Config, app *serivce.AppService) *server {
	s := &server{
		config: config,
	}
	//
	engine, run := s.createServerEngine(s.config.Port)
	s.RunServer = run
	s.engine = engine
	s.appService = app
	//
	api := s.engine.Group("/api")
	{
		collector := api.Group("/collector")
		{
			task := collector.Group("/task")
			{
				result := task.Group("/result")
				{
					repos := result.Group("/repos")
					{
						repos.POST("/by/url", s.updateStateTaskRepositoriesByURL)
					}
					issues := result.Group("/issue")
					{
						issues.POST("/by/repo", s.updateStateTaskRepositoryIssues)
					}
				}
				create := task.Group("/create")
				{
					repos := create.Group("/repos")
					{
						repos.POST("/by/url", s.createTaskRepositoriesByURL)
						repos.POST("/issues", s.createTaskRepositoriesAndIssues)
					}
					issues := create.Group("/issue")
					{
						issues.POST("/by/repos", s.createTaskRepositoriesIssues)
					}
				}
			}
		}
	}
	//s.engine.GET("/get/state", s.getState)
	//s.engine.POST("/get/repos/by/url", s.getReposByURL)
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

//
//----------------------------------------------HANDLERS (Task's create)------------------------------------------------
//

func (s *server) createTaskRepositoriesByURL(ctx *gin.Context) {
	task := new(createTaskModel.RepositoriesByURLS)
	if err := ctx.BindJSON(task); err != nil {
		runtimeinfo.LogError("request on create task [", ctx.Request.Header.Get("X-FORWARDED-FOR"), "] to ENDPOINT [", ctx.Request.URL, "] exit error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := s.appService.CreateTaskRepositoriesByURL(task.Repositories, false, nil)
	if err != nil {
		runtimeinfo.LogError("request on create task to ENDPOINT [", ctx.Request.URL, "] exit error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	runtimeinfo.LogInfo("request on create task to ENDPOINT [", ctx.Request.URL, "] status : OK")
	ctx.AbortWithStatus(http.StatusOK)
	return
}

func (s *server) createTaskRepositoriesIssues(ctx *gin.Context) {
	task := new(createTaskModel.RepositoriesByURLS)
	if err := ctx.BindJSON(task); err != nil {
		runtimeinfo.LogError("request on create task [", ctx.Request.Header.Get("X-FORWARDED-FOR"), "] to ENDPOINT [", ctx.Request.URL, "] exit error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err, nonSendTask, _ := s.appService.CreateTaskGetRepositoriesIssues(task.Repositories, true, false)
	if err != nil {
		runtimeinfo.LogError("request on create task to ENDPOINT [", ctx.Request.URL, "] exit error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	if nonSendTask != nil {
		ctx.AbortWithStatusJSON(
			http.StatusOK,
			struct {
				NonSend []string `json:"non_send"`
			}{
				NonSend: nonSendTask,
			},
		)
	}
	runtimeinfo.LogInfo("request on create task to ENDPOINT [", ctx.Request.URL, "] status : OK")
	ctx.AbortWithStatus(http.StatusOK)
	return
}

func (s *server) createTaskRepositoriesAndIssues(ctx *gin.Context) {
	task := new(createTaskModel.RepositoriesByURLS)
	if err := ctx.BindJSON(task); err != nil {
		runtimeinfo.LogError("request on create task [", ctx.Request.Header.Get("X-FORWARDED-FOR"), "] to ENDPOINT [", ctx.Request.URL, "] exit error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := s.appService.CreateTaskGetRepositoriesAndIssues(task.Repositories)
	if err != nil {
		runtimeinfo.LogError("request on create task to ENDPOINT [", ctx.Request.URL, "] exit error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	runtimeinfo.LogInfo("request on create task to ENDPOINT [", ctx.Request.URL, "] status : OK")
	ctx.AbortWithStatus(http.StatusOK)
	return
}

//
//----------------------------------------------HANDLERS (Task's update)------------------------------------------------
//

func (s *server) updateStateTaskRepositoriesByURL(ctx *gin.Context) {
	state := new(updateTaskModel.RepositoriesByURLS)
	if err := ctx.BindJSON(state); err != nil {
		runtimeinfo.LogInfo("github-collector received [", http.StatusLocked, "] response to the request to endpoint  [", ctx.Request.URL, "] with error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := s.appService.UpdateStateTaskRepositoriesByURL(state)
	if err != nil {
		runtimeinfo.LogInfo("github-collector received [", http.StatusLocked, "] response to the request to endpoint  [", ctx.Request.URL, "] with error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	runtimeinfo.LogInfo("github-collector received [", http.StatusOK, "] response to the request to endpoint  [", ctx.Request.URL, "] ")
	ctx.AbortWithStatus(http.StatusOK)
	return
}

func (s *server) updateStateTaskRepositoryIssues(ctx *gin.Context) {
	state := new(updateTaskModel.RepositoryIssues)
	if err := ctx.BindJSON(state); err != nil {
		runtimeinfo.LogInfo("github-collector received [", http.StatusLocked, "] response to the request to endpoint  [", ctx.Request.URL, "] with error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := s.appService.UpdateStateTaskRepositoryIssues(state)
	if err != nil {
		runtimeinfo.LogInfo("github-collector received [", http.StatusLocked, "] response to the request to endpoint  [", ctx.Request.URL, "] with error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	runtimeinfo.LogInfo("github-collector received [", http.StatusOK, "] response to the request to endpoint  [", ctx.Request.URL, "] ")
	ctx.AbortWithStatus(http.StatusOK)
	return
}
