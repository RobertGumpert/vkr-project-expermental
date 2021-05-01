package hashRepository

import (
	"errors"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
)

//
// NONE IMPLEMENT
//

func (storage *LocalHashStorage) HasEntities() error {
	return 	errors.New("implement me")
}

func (storage *LocalHashStorage) CreateEntities() error {
	return 	errors.New("implement me")
}

func (storage *LocalHashStorage) Migration() error {
	return 	errors.New("implement me")
}


func (storage *LocalHashStorage) AddRepository(repository *dataModel.RepositoryModel) error {
	return 	errors.New("implement me")
}

func (storage *LocalHashStorage) AddRepositories(repositories []dataModel.RepositoryModel) error {
	return 	errors.New("implement me")
}

func (storage *LocalHashStorage) GetRepositoryByName(name string) (dataModel.RepositoryModel, error) {
	return dataModel.RepositoryModel{}, nil
}

func (storage *LocalHashStorage) GetRepositoryByID(repositoryId uint) (dataModel.RepositoryModel, error) {
	return dataModel.RepositoryModel{},	errors.New("implement me")
}

func (storage *LocalHashStorage) GetAllRepositories() ([]dataModel.RepositoryModel, error) {
	return nil,	errors.New("implement me")
}

func (storage *LocalHashStorage) AddIssue(issue *dataModel.IssueModel) error {
	return 	errors.New("implement me")
}

func (storage *LocalHashStorage) AddIssues(issues []dataModel.IssueModel) error {
	return errors.New("implement me")
}

func (storage *LocalHashStorage) AddNearestIssues(nearest dataModel.NearestIssuesModel) error {
	return errors.New("implement me")
}

func (storage *LocalHashStorage) GetIssueByID(issueId uint) (dataModel.IssueModel, error) {
	return dataModel.IssueModel{}, errors.New("implement me")
}

func (storage *LocalHashStorage) GetIssueRepository(repositoryId uint) ([]dataModel.IssueModel, error) {
	return nil, errors.New("implement me")
}

func (storage *LocalHashStorage) GetNearestIssuesForIssue(issueId uint) ([]dataModel.NearestIssuesModel, error) {
	return nil, errors.New("implement me")
}

func (storage *LocalHashStorage) GetNearestIssuesForRepository(repositoryId uint) ([]dataModel.NearestIssuesModel, error) {
	return  nil, errors.New("implement me")
}

func (storage *LocalHashStorage) AddListNearestIssues(nearest []dataModel.NearestIssuesModel) error {
	panic("implement me")
}

func (storage *LocalHashStorage) GetIssuesOnlyGroupRepositories(repositoryId ...uint) ([]dataModel.IssueModel, error) {
	panic("implement me")
}

func (storage *LocalHashStorage) GetIssuesBesidesGroupRepositories(repositoryId ...uint) ([]dataModel.IssueModel, error) {
	panic("implement me")
}

func (storage *LocalHashStorage) AddNumbersIntersections(intersections []dataModel.NumberIssueIntersectionsModel) error {
	panic("implement me")
}

func (storage *LocalHashStorage) AddNumberIntersections(intersection *dataModel.NumberIssueIntersectionsModel) error {
	panic("implement me")
}

func (storage *LocalHashStorage) GetNumberIntersectionsForRepository(repositoryID uint) ([]dataModel.NumberIssueIntersectionsModel, error) {
	panic("implement me")
}

func (storage *LocalHashStorage) GetNumberIntersectionsForPair(repositoryID, comparableRepositoryID uint) (dataModel.NumberIssueIntersectionsModel, error) {
	panic("implement me")
}

func (storage *LocalHashStorage) GetNearestIssuesForPairRepositories(mainRepositoryID, secondRepositoryID uint) ([]dataModel.NearestIssuesModel, error) {
	panic("implement me")
}