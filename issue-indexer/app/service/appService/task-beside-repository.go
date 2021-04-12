package appService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"issue-indexer/app/config"
	"issue-indexer/app/service/implementComparatorRules/comparison"
	"issue-indexer/app/service/implementComparatorRules/sampling"
	"issue-indexer/app/service/issueCompator"
)

type taskCompareBesideRepository struct {
	taskManager     itask.IManager
	comparator      *issueCompator.Comparator
	config          *config.Config
	samplingRules   *sampling.ImplementRules
	comparisonRules *comparison.ImplementRules
}

func newTaskCompareBesideRepository(
	taskManager itask.IManager,
	comparator *issueCompator.Comparator,
	config *config.Config,
	samplingRules *sampling.ImplementRules,
	comparisonRules *comparison.ImplementRules,
) *taskCompareBesideRepository {
	return &taskCompareBesideRepository{
		taskManager:     taskManager,
		comparator:      comparator,
		config:          config,
		samplingRules:   samplingRules,
		comparisonRules: comparisonRules,
	}
}

func (facade *taskCompareBesideRepository) CreateTask(taskKey string, repositoryID uint, returnResult issueCompator.ReturnResult) (task itask.ITask, err error) {
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
		CrossingThreshold: facade.config.MinimumTextCompletenessThreshold,
	}
	task, err = facade.taskManager.CreateTask(
		compareBesideRepository,
		taskKey,
		nil,
		nil,
		conditionSampling,
		facade.EventRunTask,
		facade.EventUpdateTaskState,
	)
	if err != nil {
		return nil, err
	}
	rules = issueCompator.NewCompareRules(
		repositoryID,
		int64(facade.config.MaxCountThreads),
		facade.samplingRules.IssuesOnlyFromGroupRepositories,
		facade.comparisonRules.CompareBodyAfterCompareTitles,
		returnResult,
		conditionComparison,
		conditionSampling,
	)
	result = issueCompator.NewCompareResult(task)
	task.GetState().SetSendContext(&sendToComparatorContext{
		rules:  rules,
		result: result,
	})
	task.GetState().SetUpdateContext(result)
	return task, nil
}

func (facade *taskCompareBesideRepository) EventManageTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	deleteTasks = map[string]struct{}{task.GetKey(): {}}
	task.GetState().SetCustomFields(
		&sendToGateContext{
			endpoint: facade.config.GithubGateEndpoints.SendResultTaskCompareBeside,
			taskKey:  task.GetKey(),
			jsonBody: &jsonSendToGateCompareBeside{
				ExecutionTaskStatus: jsonExecutionTaskStatus{
					TaskKey:       task.GetKey(),
					TaskCompleted: true,
				},
			},
		},
	)
	return deleteTasks
}

func (facade *taskCompareBesideRepository) EventRunTask(task itask.ITask) (doTaskAsDefer, sendToErrorChannel bool, err error) {
	send := task.GetState().GetSendContext().(*sendToComparatorContext)
	err = facade.comparator.DOCompare(
		send.GetRules(),
		send.GetResult(),
	)
	if err != nil {
		return true, true, err
	}
	return false, false, nil
}

func (facade *taskCompareBesideRepository) EventUpdateTaskState(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	update := somethingUpdateContext.(*issueCompator.CompareResult)
	if update.GetErr() != nil {
		return update.GetErr(), true
	}
	task.GetState().SetCompleted(true)
	return nil, false
}
