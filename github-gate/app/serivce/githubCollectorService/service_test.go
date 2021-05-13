package githubCollectorService

import (
	"github-gate/app/config"
	"github-gate/app/models/customFieldsModel"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/gotasker/tasker"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/gin-gonic/gin"
	"log"
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

func createTask(taskKey string) (task itask.ITask, err error) {
	return taskManager.CreateTask(
		0,
		taskKey,
		make([]dataModel.RepositoryModel, 0),
		make([]dataModel.RepositoryModel, 0),
		&customFieldsModel.Model{
			TaskType: 0,
			Fields:   chanResult,
		},
		nil, nil,
	)
}


func initDependents() (*config.Config, repository.IRepository, *gin.Engine) {
	CONFIG := config.NewConfig().ReadWithPath("C:/VKR/vkr-project-expermental/github-gate/data/config/config.json")
	POSTGRES := repository.NewSQLRepository(
		repository.SQLCreateConnection(
			repository.TypeStoragePostgres,
			repository.DSNPostgres,
			nil,
			CONFIG.Postgres.Username,
			CONFIG.Postgres.Password,
			CONFIG.Postgres.DbName,
			CONFIG.Postgres.Port,
			CONFIG.Postgres.Ssl,
		),
	)
	SERVER := gin.Default()
	return CONFIG, POSTGRES, SERVER
}

func TestTaskForCollectorDownloadRepositoryFlow(t *testing.T) {
	c, p, s := initDependents()
	go scan()
	//
	service := NewCollectorService(
		p,
		c,
		s,
	)
	task, err := createTask("test_0")
	if err != nil {
		t.Fatal(err)
	}
	err = service.CreateTaskRepositoriesByName(
		task,
		dataModel.RepositoryModel{
			Name:  "gin",
			Owner: "gin-gonic",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Run(":" + c.Port)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTaskForCollectorDownloadRepositoryAndContainingKeywordFlow(t *testing.T) {
	c, p, s := initDependents()
	go scan()
	//
	service := NewCollectorService(
		p,
		c,
		s,
	)
	task, err := createTask("test_0")
	if err != nil {
		t.Fatal(err)
	}
	err = service.CreateTaskRepositoryAndRepositoriesContainingKeyword(
		task,
		dataModel.RepositoryModel{
			Name:  "Spectator",
			Owner: "y2k",
		},
		"telegram",
	)
	if err != nil {
		t.Fatal(err)
	}
	err = s.Run(":" + c.Port)
	if err != nil {
		t.Fatal(err)
	}
}