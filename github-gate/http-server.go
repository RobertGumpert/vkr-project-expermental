package main

import (
	"github-gate/app/config"
	//"github.com/RobertGumpert/vkr-pckg/tree/main/runtimeinfo"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
)

type server struct {
	config     *config.Config
	engine     *gin.Engine
	RunServer  func()
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
	//api := s.engine.Group("/api")
	//{
	//	collector := api.Group("/collector")
	//	{
	//		task := collector.Group("/task")
	//		{
	//			result := task.Group("/result")
	//			{
	//				repos := result.Group("/repos")
	//				{
	//					repos.POST("/by/url", s.updateStateTaskRepositoriesByURL)
	//				}
	//				issues := result.Group("/issue")
	//				{
	//					issues.POST("/by/repo", s.updateStateTaskRepositoryIssues)
	//				}
	//			}
	//			create := task.Group("/create")
	//			{
	//				repos := create.Group("/repos")
	//				{
	//					repos.POST("/by/url", s.createTaskRepositoriesByURL)
	//					repos.POST("/issues", s.createTaskRepositoriesAndIssues)
	//				}
	//				issues := create.Group("/issue")
	//				{
	//					issues.POST("/by/repos", s.createTaskRepositoriesIssues)
	//				}
	//			}
	//		}
	//	}
	//}
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
			//runtimeinfo.LogFatal(err)
		}
	}
}

//
//----------------------------------------------HANDLERS (Task's create)------------------------------------------------
//

