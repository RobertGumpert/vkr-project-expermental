package appService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"issue-indexer/app/config"
	"issue-indexer/app/service/implementComparatorRules/comparison"
	"issue-indexer/app/service/implementComparatorRules/sampling"
	"issue-indexer/app/service/issueCompator"
	"time"
)

type tasksFacade struct {
	taskManager     itask.IManager
	comparator      *issueCompator.Comparator
	config          *config.Config
	samplingRules   *sampling.ImplementRules
	comparisonRules *comparison.ImplementRules
	//
	taskGroup  *taskCompareWithGroupRepositories
	taskBeside *taskCompareBesideRepository
}

func newTasksFacade(taskManager itask.IManager, config *config.Config, db repository.IRepository) *tasksFacade {
	facade := new(tasksFacade)
	facade.comparator = issueCompator.NewComparator(db)
	facade.samplingRules = sampling.NewSampler(
		db,
		time.Duration(config.MaximumDurationDatabaseQueryInMinutes)*time.Minute,
	)
	facade.comparisonRules = comparison.NewImplementRules()
	facade.taskGroup = newTaskCompareWithGroupRepositories(
		taskManager,
		facade.comparator,
		config,
		facade.samplingRules,
		facade.comparisonRules,
	)
	facade.taskBeside = newTaskCompareBesideRepository(
		taskManager,
		facade.comparator,
		config,
		facade.samplingRules,
		facade.comparisonRules,
	)
	return facade
}

func (t *tasksFacade) GetTaskCompareBesideRepository() *taskCompareBesideRepository {
	return t.taskBeside
}

func (t *tasksFacade) GetTaskCompareGroupRepositories() *taskCompareWithGroupRepositories {
	return t.taskGroup
}
