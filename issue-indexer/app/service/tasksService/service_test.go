package tasksService

import (
	"encoding/json"
	"issue-indexer/app/config"
	"issue-indexer/app/models/createTaskModel"
	"issue-indexer/app/models/dataModel"
	"issue-indexer/app/repository"
	"issue-indexer/pckg/runtimeinfo"
	"issue-indexer/pckg/textPreprocessing/textDictionary"
	"issue-indexer/pckg/textPreprocessing/textVectorized"
	"log"
	"runtime"
	"testing"
	"time"
)

var storageProvider *repository.ApplicationStorageProvider

func connect() repository.IRepositoriesStorage {
	storageProvider = repository.SQLCreateConnection(
		repository.TypeStoragePostgres,
		repository.DSNPostgres,
		nil,
		"postgres",
		"toster123",
		"vkr-db",
		"5432",
		"disable",
	)
	sqlRepository := repository.NewSQLRepository(
		storageProvider,
	)
	return sqlRepository
}

func createFakeFrequencyJSON(str string) ([]byte, error) {
	slice := textDictionary.TextTransformToFeaturesSlice(str)
	dict := textVectorized.GetFrequencyMap(slice)
	m := make(map[string]float64, 0)
	for item := range dict.IterBuffered() {
		m[item.Key] = item.Val.(float64)
	}
	obj := &dataModel.TitleFrequencyJSON{Dictionary: m}
	return json.Marshal(obj)
}

func createFakeData(db repository.IRepositoriesStorage) []*dataModel.Repository {
	titleA := "Feature Request: Warnings for missing Aria properties in debug mode"
	titleB := "Feature Request: properties in debug mode"
	btsA, err := createFakeFrequencyJSON(titleA)
	btsB, err := createFakeFrequencyJSON(titleB)
	if err != nil {
		panic(err)
	}
	//
	repositories := []*dataModel.Repository{
		{
			URL:         "a",
			Name:        "a",
			Owner:       "a",
			Topics:      []string{"a", "a", "a"},
			Description: "a",
		},
		{
			URL:         "b",
			Name:        "b",
			Owner:       "b",
			Topics:      []string{"b", "b", "b"},
			Description: "b",
		},
		{
			URL:         "c",
			Name:        "c",
			Owner:       "c",
			Topics:      []string{"c", "c", "c"},
			Description: "c",
		},
	}
	err = db.AddRepositories(repositories)
	if err != nil {
		runtimeinfo.LogError(err)
		log.Fatal()
	}
	for i := 0; i < 3; i++ {
		err = db.AddIssues([]*dataModel.Issue{
			{
				RepositoryID:       repositories[0].ID,
				Number:             i,
				Title:              titleA,
				TitleDictionary:    []string{"a"},
				TitleFrequencyJSON: btsA,
			},
			{
				RepositoryID:       repositories[1].ID,
				Number:             i,
				Title:              titleA,
				TitleDictionary:    []string{"b"},
				TitleFrequencyJSON: btsB,
			},
			{
				RepositoryID:       repositories[2].ID,
				Number:             i,
				Title:              titleA,
				TitleDictionary:    []string{"c"},
				TitleFrequencyJSON: btsA,
			},
		})
		if err != nil {
			runtimeinfo.LogError(err)
			log.Fatal()
		}
	}
	return repositories
}

func TestTruncate(t *testing.T) {
	_ = connect()
	storageProvider.SqlDB.Exec("TRUNCATE TABLE repositories CASCADE")
	storageProvider.SqlDB.Exec("TRUNCATE TABLE issues CASCADE")
}

func TestAddFlow(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	db := connect()
	repositories := createFakeData(db)
	c := &config.Config{
		MaxCountRunnableTasks:            2,
		MaxCountThreads:                  5,
		MinimumTextCompletenessThreshold: 50.0,
	}
	service := NewTasksService(c, db)
	for i := 0; i < len(repositories); i++ {
		ids := make([]uint, 0)
		for j := 0; j < len(repositories); j++ {
			if i == j {
				continue
			}
			ids = append(ids, repositories[j].ID)
		}
		go func(i int, service *TasksService, repositories []*dataModel.Repository, ids []uint) {
			err := service.CreateTaskCompareIssuesInPairs(&createTaskModel.CreateTaskCompareIssuesInPairs{
				TaskKey:                   repositories[i].Name,
				ComparableRepositoryID:    repositories[i].ID,
				CompareWithRepositoriesID: ids,
			})
			if err != nil {
				t.Fatal(err)
			}
		}(i, service, repositories, ids)
	}
	time.Sleep(1 * time.Hour)
}
