package appService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/gotasker/tasker"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"issue-indexer/app/config"
	"issue-indexer/app/service/implementComparatorRules/comparison"
	"issue-indexer/app/service/implementComparatorRules/sampling"
	"issue-indexer/app/service/issueCompator"
	"strconv"
	"strings"
	"time"
)

type AppService struct {
	taskSteward     *tasker.Steward
	samplingRules   *sampling.ImplementRules
	comparisonRules *comparison.ImplementRules
	comparator      *issueCompator.Comparator
	db              repository.IRepository
	config          *config.Config
}

func NewAppService(db repository.IRepository, config *config.Config) *AppService {
	service := new(AppService)
	service.db = db
	service.config = config
	service.comparator = issueCompator.NewComparator(db)
	service.taskSteward = tasker.NewSteward(
		int64(config.MaxCountRunnableTasks),
		10*time.Second,
		nil,
	)
	return service
}

func (service *AppService) compareOnlyBesideRepository(repositoryID uint) (task itask.ITask, err error) {
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

func (service *AppService) compareOnlyWithGroupRepositories(repositoryID uint, comparableRepositoriesID []uint) (task itask.ITask, err error) {
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

func (service *AppService) returnResultFromComparator(result *issueCompator.CompareResult) {
	task := result.GetIdentifier().(itask.ITask)
	task.GetState().SetCompleted(true)
	if err := result.GetErr(); err != nil {
		task.GetState().SetError(err)
		runtimeinfo.LogError(err)
	}
	if err := service.taskSteward.UpdateTask(
		task.GetKey(),
		task.GetState().GetSendContext(),
	); err != nil {
		runtimeinfo.LogError(err)
	}
}
