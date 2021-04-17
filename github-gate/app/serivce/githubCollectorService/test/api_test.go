package test

import (
	"github-gate/app/models/customFieldsModel"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/gotasker/tasker"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"testing"
)

var (
	chanResult  = make(chan itask.ITask)
	taskManager = tasker.NewManager(
		tasker.SetBaseOptions(1000, nil),
	)
)

func scan() {
	for task := range chanResult {
		updateContext := task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		for _, model := range updateContext {
			log.Println("Add repo. ID: ", model.ID)
		}
	}
}

func createTask() (task itask.ITask, err error) {
	return taskManager.CreateTask(
		0, "", nil,
		make([]dataModel.RepositoryModel, 0),
		&customFieldsModel.Model{
			TaskType: 0,
			Fields:   chanResult,
		},
		nil, nil,
	)
}

func TestApiFlow(t *testing.T) {
	type (
		repo struct {
			Name  string `json:"name"`
			Owner string `json:"owner"`
		}
		repositories struct {
			Repos []repo `json:"repos"`
		}
		keyword struct {
			Keyword string `json:"keyword"`
		}
		repoAndKeyword struct {
			Repo    repo   `json:"repo"`
			Keyword string `json:"keyword"`
		}
	)
	var (
		c               = createFakeConfig()
		service, server = createFakeTaskService(c)
	)
	go scan()
	tc := server.Group("/test/collector")
	{
		tc.POST("/by/name", func(context *gin.Context) {
			state := new(repositories)
			if err := context.BindJSON(state); err != nil {
				runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
				context.AbortWithStatus(http.StatusLocked)
				return
			}
			task, err := createTask()
			if err != nil {
				runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
				context.AbortWithStatus(http.StatusLocked)
				return
			}
			models := func() []dataModel.RepositoryModel {
				m := make([]dataModel.RepositoryModel, 0)
				for _, r := range state.Repos {
					m = append(
						m,
						dataModel.RepositoryModel{
							Name:  r.Name,
							Owner: r.Owner,
						},
					)
				}
				return m
			}()
			task.GetState().SetSendContext(models)
			err = service.CreateTaskRepositoriesByName(
				task,
				models...,
			)
			if err != nil {
				runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
				context.AbortWithStatus(http.StatusLocked)
				return
			}
			context.AbortWithStatus(http.StatusOK)
			return
		})
		tc.POST("/by/keyword", func(context *gin.Context) {
			state := new(keyword)
			if err := context.BindJSON(state); err != nil {
				runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
				context.AbortWithStatus(http.StatusLocked)
				return
			}
			task, err := createTask()
			if err != nil {
				runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
				context.AbortWithStatus(http.StatusLocked)
				return
			}
			err = service.CreateTaskRepositoriesByKeyword(
				task,
				state.Keyword,
			)
			if err != nil {
				runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
				context.AbortWithStatus(http.StatusLocked)
				return
			}
			context.AbortWithStatus(http.StatusOK)
			return
		})
		tc.POST("/repository/and/repositories/by/keyword", func(context *gin.Context) {
			state := new(repoAndKeyword)
			if err := context.BindJSON(state); err != nil {
				runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
				context.AbortWithStatus(http.StatusLocked)
				return
			}
			task, err := createTask()
			if err != nil {
				runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
				context.AbortWithStatus(http.StatusLocked)
				return
			}
			err = service.CreateTaskRepositoryAndRepositoriesContainingKeyword(
				task,
				dataModel.RepositoryModel{Name: state.Repo.Name, Owner: state.Repo.Owner},
				state.Keyword,
			)
			if err != nil {
				runtimeinfo.LogError("(RESP. TO: -> GITHUB-COLLECTOR) JSON UNMARSHAL COMPLETED WITH ERROR: ", err)
				context.AbortWithStatus(http.StatusLocked)
				return
			}
			context.AbortWithStatus(http.StatusOK)
			return
		})
	}

	err := server.Run(":" + c.Port)
	if err != nil {
		t.Fatal(err)
	}
}
