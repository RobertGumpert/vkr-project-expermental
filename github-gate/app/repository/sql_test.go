package repository

import (
	"github-gate/app/models/dataModel"
	"github-gate/pckg/runtimeinfo"
	"testing"
)

func connect() IRepositoriesStorage {
	sqlRepository := NewSQLRepository(
		SQLCreateConnection(
			TypeStoragePostgres,
			DSNPostgres,
			nil,
			"postgres",
			"toster123",
			"vkr-db",
			"5432",
			"disable",
		),
	)
	return sqlRepository
}

func TestAddFlow(t *testing.T) {
	db := connect()
	//
	repository := &dataModel.Repository{
		URL:         "a",
		Name:        "a",
		Owner:       "a",
		Topics:      []string{"a", "a", "a"},
		Description: "a",
	}
	repositories := []*dataModel.Repository{
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
	//
	err := db.AddRepository(repository)
	if err != nil {
		runtimeinfo.LogError(err)
		t.Fatal()
	}
	//
	err = db.AddRepositories(repositories)
	if err != nil {
		runtimeinfo.LogError(err)
		t.Fatal()
	}
	//
	err = db.AddIssues([]*dataModel.Issue{
		{
			RepositoryID: repository.ID,
			Number:       1,
			URL:          "a",
			Title:        "a",
			State:        "a",
			Body:         "a",
		},
	})
	if err != nil {
		runtimeinfo.LogError(err)
		t.Fatal()
	}
	//
	issues := make([]*dataModel.Issue, 0)
	for index, repo := range repositories {
		issues = append(issues, &dataModel.Issue{
			RepositoryID: repo.ID,
			Number:       index,
			URL:          repo.URL,
			Title:        repo.URL,
			State:        repo.URL,
			Body:         repo.URL,
		})
	}
	err = db.AddIssues(issues)
	if err != nil {
		runtimeinfo.LogError(err)
		t.Fatal()
	}
	//
	err = db.CloseConnection()
	if err != nil {
		runtimeinfo.LogError(err)
		t.Fatal()
	}
	runtimeinfo.LogInfo("Ok")
}
