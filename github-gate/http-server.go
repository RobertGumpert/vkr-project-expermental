package main

import (
	"github-gate/pckg/runtimeinfo"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
						repos.POST("/by/url", s.updateStateTaskReposByURL)
					}
					issues := result.Group("/issue")
					{
						issues.POST("/by/url", s.updateStateTaskIssueByRepo)
					}
				}
				create := task.Group("/create")
				{
					repos := create.Group("/repos")
					{
						repos.POST("/by/url", s.createTaskReposByURL)
					}
					issues := create.Group("/issue")
					{
						issues.POST("/by/url", s.createTaskIssueByRepo)
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

func (s *server) createTaskReposByURL(ctx *gin.Context) {
	task := new(CreateTaskRepoByURLS)
	if err := ctx.BindJSON(task); err != nil {
		runtimeinfo.LogError("request on create task [", ctx.Request.Header.Get("X-FORWARDED-FOR"), "] to ENDPOINT [", ctx.Request.URL, "] exit error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := s.appService.CreateTaskReposByURL(task)
	if err != nil {
		runtimeinfo.LogError("request on create task [", ctx.Request.Header.Get("X-FORWARDED-FOR"), "] to ENDPOINT [", ctx.Request.URL, "] exit error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	runtimeinfo.LogInfo("request on create task [", ctx.Request.Header.Get("X-FORWARDED-FOR"), "] to ENDPOINT [", ctx.Request.URL, "] status : OK")
	ctx.AbortWithStatus(http.StatusOK)
	return
}

func (s *server) createTaskIssueByRepo(ctx *gin.Context) {

}


//
//----------------------------------------------HANDLERS (Task's update)------------------------------------------------
//

func (s *server) updateStateTaskReposByURL(ctx *gin.Context) {
	state := new(UpdateTaskReposByURLS)
	if err := ctx.BindJSON(state); err != nil {
		runtimeinfo.LogError("request on update task [", ctx.Request.Header.Get("X-FORWARDED-FOR"), "] to ENDPOINT [", ctx.Request.URL, "] exit error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	err := s.appService.UpdateStateTaskReposByURL(state)
	if err != nil {
		runtimeinfo.LogError("request on update task [", ctx.Request.Header.Get("X-FORWARDED-FOR"), "] to ENDPOINT [", ctx.Request.URL, "] exit error: ", err)
		ctx.AbortWithStatus(http.StatusLocked)
		return
	}
	runtimeinfo.LogInfo("request on update task [", ctx.Request.Header.Get("X-FORWARDED-FOR"), "] to ENDPOINT [", ctx.Request.URL, "] status : OK")
	ctx.AbortWithStatus(http.StatusOK)
	return
}

func (s *server) updateStateTaskIssueByRepo(ctx *gin.Context) {

}
