package appService

import (
	"github-collector/app/config"
	"github-collector/pckg/runtimeinfo"
	"testing"
)

func TestDownloadRepositoryFlow(t *testing.T) {
	fakeChannel := make(chan bool)
	c := config.NewConfig().ReadWithPath("C:/VKR/vkr-project-expermental/github-collector/data/config/config.json")
	service := NewAppService(c)
	err := service.CreateTaskRepositoriesDescriptions(&JsonCreateTaskRepositoriesDescriptions{
		TaskKey: "test_0",
		Repositories: []JsonRepository{
			{
				Name:  "gin",
				Owner: "gin-gonic",
			},
			{
				Name:  "vue",
				Owner: "vuejs",
			},
		},
	})
	if err != nil {
		runtimeinfo.LogError(err)
	}
	fakeChannel <- true
}

func TestDownloadRepositoryIssuesFlow(t *testing.T) {
	fakeChannel := make(chan bool)
	c := config.NewConfig().ReadWithPath("C:/VKR/vkr-project-expermental/github-collector/data/config/config.json")
	service := NewAppService(c)
	err := service.CreateTaskRepositoryIssues(&JsonCreateTaskRepositoryIssues{
		TaskKey: "test_0",
		Repository: JsonRepository{
			Name:  "visx",
			Owner: "airbnb",
		},
	})
	if err != nil {
		runtimeinfo.LogError(err)
	}
	fakeChannel <- true
}

func TestDownloadRepositoryByKeywordFlow(t *testing.T) {
	fakeChannel := make(chan bool)
	c := config.NewConfig().ReadWithPath("C:/VKR/vkr-project-expermental/github-collector/data/config/config.json")
	service := NewAppService(c)
	err := service.CreateTaskRepositoriesByKeyWord(&JsonCreateTaskRepositoriesByKeyWord{
		TaskKey: "test_0",
		KeyWord: "vpn",
	})
	if err != nil {
		runtimeinfo.LogError(err)
	}

	//err = service.CreateTaskRepositoriesByKeyWord(&JsonCreateTaskRepositoriesByKeyWord{
	//	TaskKey: "test_1",
	//	KeyWord: "vpn",
	//})
	//if err != nil {
	//	runtimeinfo.LogError(err)
	//}
	fakeChannel <- true
}
