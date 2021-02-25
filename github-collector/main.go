package main

import(
	githubRequest "github-collector/pckg/github-api/github-request"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

var(
	APP *AppService
	GITHUBClient *githubRequest.GithubClient
)

type GetRepoJSONModel struct {
	URLS []string `json:"urls"`
}

func main() {
	engine, run := getServer()
	client, err := githubRequest.NewGithubClient("")
	if err != nil {
		log.Fatal(err)
	}
	APP = NewAppService(client)
	//
	engine.GET("/get/state", getState)
	engine.POST("/get/repos", getRepos)
	//
	run()
}

func getState(ctx *gin.Context) {
	if GITHUBClient.WaitRateLimitsReset {
		ctx.AbortWithStatus(http.StatusLocked)
		return
	} else {
		ctx.AbortWithStatus(http.StatusOK)
		return
	}
}

func getRepos(ctx *gin.Context) {
	data := new(GetRepoJSONModel)
	if err := ctx.BindJSON(data); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
}

func getServer(port ...string) (*gin.Engine, func()) {
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