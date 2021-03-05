package repository

import "github-gate/app/models/dataModel"

type IStorage interface {
	HasEntities() error
	CreateEntities() error
	Migration() error
	CloseConnection() error
}

type IRepositoriesStorage interface {
	IStorage
	//
	AddRepository(repository *dataModel.Repository) error
	AddRepositories(repositories []*dataModel.Repository) error
	//
	AddIssues(issues []*dataModel.Issue) error
}
