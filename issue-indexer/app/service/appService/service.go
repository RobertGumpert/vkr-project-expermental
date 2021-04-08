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

func (service *AppService) returnResultToGate(task itask.ITask, result *issueCompator.CompareResult) {
	if !task.GetState().IsCompleted() {
		runtimeinfo.LogError("TASK NOT COMPLETED. ")
	}
}