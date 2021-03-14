package tasksService

import "issue-indexer/app/models/dataModel"

func (service *TasksService) readIssuesForCompareInPairs(comparableRepositoryID uint, compareWithRepositoriesID []uint) (comparable, compareWith []dataModel.Issue, err error) {
	comparable, err = service.db.ListIssuesRepository(comparableRepositoryID)
	if err != nil {
		return comparable, compareWith, err
	}
	compareWith, err = service.db.ListIssuesInRepositories(compareWithRepositoriesID)
	if err != nil {
		return comparable, compareWith, err
	}
	return comparable, compareWith, nil
}
