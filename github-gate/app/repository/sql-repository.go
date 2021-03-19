package repository

import (
	"errors"
	"github-gate/app/models/dataModel"
	"github-gate/pckg/runtimeinfo"
)

type SQLRepository struct {
	storage *ApplicationStorageProvider
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

func (s *SQLRepository) CreateEntities() error {
	db := s.storage.SqlDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
		}
	}()
	if err := db.Migrator().CreateTable(
		&dataModel.Repository{},
		&dataModel.Issue{},
		&dataModel.NearestIssues{},
	); err != nil {
		db.Rollback()
		return err
	}
	return db.Commit().Error
}

func (s *SQLRepository) Migration() error {
	db := s.storage.SqlDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
		}
	}()
	if err := db.AutoMigrate(
		&dataModel.Repository{},
		&dataModel.Issue{},
		&dataModel.NearestIssues{},
	); err != nil {
		db.Rollback()
		return err
	}
	return db.Commit().Error
}

func (s *SQLRepository) CloseConnection() error {
	db, err := s.storage.SqlDB.DB()
	if err != nil {
		return err
	}
	err = db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLRepository) GetRepositoryByName(name string) (dataModel.Repository, error) {
	var repository dataModel.Repository
	if err := s.storage.SqlDB.Where("name = ?", name).First(&repository).Error; err != nil {
		return repository, err
	}
	return repository, nil
}

func (s *SQLRepository) AddRepository(repository dataModel.Repository) error {
	tx := s.storage.SqlDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Create(&repository).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *SQLRepository) AddRepositories(repositories []dataModel.Repository) error {
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

func (s *SQLRepository) AddIssues(issues []dataModel.Issue) error {
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

func NewSQLRepository(storage *ApplicationStorageProvider) *SQLRepository {
	repository := &SQLRepository{storage: storage}
	err := repository.HasEntities()
	if err != nil {
		err := repository.CreateEntities()
		if err != nil {
			runtimeinfo.LogFatal(err)
		}
	}
	err = repository.Migration()
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	return repository
}
