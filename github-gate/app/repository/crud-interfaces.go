package repository

import "github-gate/app/models/dataModel"

type IStorage interface {
	HasEntities() error
	CreateEntities() error
	Migration() error
	CloseConnection() error
}

type IRepository interface {
	IStorage
	GetRepositoryByName(name string) (dataModel.Repository, error)
	AddRepository(repository dataModel.Repository) error
	AddRepositories(repositories []dataModel.Repository) error
	AddIssues(issues []dataModel.Issue) error
	ListIssuesRepository(id uint) ([]dataModel.Issue, error)
	AddNearestIssues(nearestIssues dataModel.NearestIssues) error
}
