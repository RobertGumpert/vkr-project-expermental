package appService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"issue-indexer/app/service/implementComparatorRules/comparison"
	"issue-indexer/app/service/implementComparatorRules/sampling"
	"issue-indexer/app/service/issueCompator"
	"strconv"
	"strings"
)

func (service *AppService) createTaskCompareOnlyBesideRepository(repositoryID uint) (task itask.ITask, err error) {
	var (
		rules               *issueCompator.CompareRules
		result              *issueCompator.CompareResult
		conditionSampling   *sampling.ConditionIssuesBesidesRepository
		conditionComparison *comparison.ConditionIntersections
	)
	conditionSampling = &sampling.ConditionIssuesBesidesRepository{
		RepositoryID: repositoryID,
	}
	conditionComparison = &comparison.ConditionIntersections{
		CrossingThreshold: service.config.MinimumTextCompletenessThreshold,
	}
	task, err = service.taskSteward.CreateTask(
		compareWithGroupRepositories,
		strings.Join([]string{
			"beside-",
			strconv.Itoa(int(repositoryID)),
		}, "-"),
		nil,
		nil,
		conditionSampling,
		service.eventRunTask,
		service.eventUpdateTaskState,
	)()
	if err != nil {
		return nil, err
	}
	rules = issueCompator.NewCompareRules(
		repositoryID,
		int64(service.config.MaxCountThreads),
		service.samplingRules.IssuesOnlyFromGroupRepositories,
		service.comparisonRules.CompareTitlesWithConditionIntersection,
		service.returnResultFromComparator,
		conditionComparison,
		conditionSampling,
	)
	result = issueCompator.NewCompareResult(task)
	task.GetState().SetSendContext(&sendContext{
		rules:  rules,
		result: result,
	})
	task.GetState().SetUpdateContext(result)
	return task, nil
}

func (service *AppService) createTaskCompareOnlyWithGroupRepositories(repositoryID uint, comparableRepositoriesID []uint) (task itask.ITask, err error) {
	var (
		rules               *issueCompator.CompareRules
		result              *issueCompator.CompareResult
		conditionSampling   *sampling.ConditionIssuesFromGroupRepository
		conditionComparison *comparison.ConditionIntersections
	)
	conditionSampling = &sampling.ConditionIssuesFromGroupRepository{
		RepositoryID:      repositoryID,
		GroupRepositories: comparableRepositoriesID,
	}
	conditionComparison = &comparison.ConditionIntersections{
		CrossingThreshold: service.config.MinimumTextCompletenessThreshold,
	}
	task, err = service.taskSteward.CreateTask(
		compareWithGroupRepositories,
		strings.Join([]string{
			"with-group",
			strconv.Itoa(int(repositoryID)),
		}, "-"),
		nil,
		nil,
		conditionSampling,
		service.eventRunTask,
		service.eventUpdateTaskState,
	)()
	if err != nil {
		return nil, err
	}
	rules = issueCompator.NewCompareRules(
		repositoryID,
		int64(service.config.MaxCountThreads),
		service.samplingRules.IssuesOnlyFromGroupRepositories,
		service.comparisonRules.CompareTitlesWithConditionIntersection,
		service.returnResultFromComparator,
		conditionComparison,
		conditionSampling,
	)
	result = issueCompator.NewCompareResult(task)
	task.GetState().SetSendContext(&sendContext{
		rules:  rules,
		result: result,
	})
	task.GetState().SetUpdateContext(result)
	return task, nil
}

