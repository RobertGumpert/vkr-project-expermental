package repository

import "issue-indexer/app/models/dataModel"

type IStorage interface {
	HasEntities() error
	CloseConnection() error
}

type IRepositoriesStorage interface {
	IStorage
	//
	AddRepository(repository *dataModel.Repository) error
	AddRepositories(repositories []*dataModel.Repository) error
	AddIssues(issues []*dataModel.Issue) error
	//
	ListIssuesRepository(id uint) ([]dataModel.Issue, error)
	ListIssuesInRepositories(id []uint) ([]dataModel.Issue, error)
	AddNearestIssues(nearestIssues dataModel.NearestIssues) error
}
