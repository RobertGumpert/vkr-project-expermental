package repository

import (
	"errors"
	"issue-indexer/app/models/dataModel"
	"issue-indexer/pckg/runtimeinfo"
)

type SQLRepository struct {
	storage *ApplicationStorageProvider
}

func (s *SQLRepository) CloseConnection() error {
	panic("implement me")
}

func (s *SQLRepository) HasEntities() error {
	db := s.storage.SqlDB.Begin()
	entities := []interface{}{
		&dataModel.Repository{},
		&dataModel.Issue{},
		&dataModel.NearestIssues{},
	}
	for _, entity := range entities {
		if exist := db.Migrator().HasTable(entity); !exist {
			return errors.New("Non exist table. ")
		}
	}
	return nil
}


func (s *SQLRepository) AddRepository(repository *dataModel.Repository) error {
	tx := s.storage.SqlDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Create(repository).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *SQLRepository) AddRepositories(repositories []*dataModel.Repository) error {
	tx := s.storage.SqlDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Create(&repositories).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *SQLRepository) AddIssues(issues []*dataModel.Issue) error {
	tx := s.storage.SqlDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Create(&issues).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *SQLRepository) ListIssuesRepository(id uint) ([]dataModel.Issue, error) {
	var issues []dataModel.Issue
	if err := s.storage.SqlDB.Where("repository_id = ?", id).Find(&issues).Error; err != nil {
		return issues, err
	}
	return issues, nil
}

func (s *SQLRepository) AddNearestIssues(nearestIssues dataModel.NearestIssues) error {
	tx := s.storage.SqlDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Create(&nearestIssues).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *SQLRepository) ListIssuesInRepositories(id []uint) ([]dataModel.Issue, error) {
	var issues []dataModel.Issue
	if err := s.storage.SqlDB.Where("repository_id IN ?", id).Find(&issues).Error; err != nil {
		return issues, err
	}
	return issues, nil
}

func NewSQLRepository(storage *ApplicationStorageProvider) *SQLRepository {
	repository := &SQLRepository{storage: storage}
	err := repository.HasEntities()
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	return repository
}
