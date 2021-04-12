package sampling

import (
	"context"
	"errors"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"github.com/RobertGumpert/vkr-pckg/repository"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"issue-indexer/app/service/issueCompator"
	"time"
)

type ImplementRules struct {
	timeOutContext time.Duration
	db             repository.IRepository
}

func NewSampler(db repository.IRepository, timeOutContext time.Duration) *ImplementRules {
	return &ImplementRules{
		db:             db,
		timeOutContext: timeOutContext,
	}
}

func (implement *ImplementRules) IssuesOnlyFromGroupRepositories(rules *issueCompator.CompareRules) (toCompare, doNotCompare []dataModel.IssueModel, err error) {
	var (
		ctx, cancel          = context.WithTimeout(context.Background(), implement.timeOutContext)
		condition            = rules.GetSamplingCondition().(*ConditionIssuesFromGroupRepository)
		channelGettingIssues = make(chan []dataModel.IssueModel)
	)
	defer cancel()
	go implement.getIssueOnlyFromGroupRepositories(
		channelGettingIssues,
		condition.GroupRepositories...,
	)
	select {
	case <-ctx.Done():
		return nil, nil, errors.New("READ FROM DB VERY SLOW. ")
	case toCompare = <-channelGettingIssues:
		return toCompare, nil, nil
	}
}

func (implement *ImplementRules) IssuesOnlyBesidesRepository(rules *issueCompator.CompareRules) (toCompare, doNotCompare []dataModel.IssueModel, err error) {
	var (
		ctx, cancel          = context.WithTimeout(context.Background(), implement.timeOutContext)
		condition            = rules.GetSamplingCondition().(ConditionIssuesBesidesRepository)
		channelGettingIssues = make(chan []dataModel.IssueModel)
	)
	defer cancel()
	go implement.getIssueOnlyBesidesGroupRepositories(
		channelGettingIssues,
		condition.RepositoryID,
	)
	select {
	case <-ctx.Done():
		return nil, nil, errors.New("READ FROM DB VERY SLOW. ")
	case toCompare = <-channelGettingIssues:
		return toCompare, nil, nil
	}
}

func (implement *ImplementRules) getIssueOnlyBesidesGroupRepositories(result chan []dataModel.IssueModel, repositoryID ...uint) {
	var (
		issues, err = implement.db.GetIssuesBesidesGroupRepositories(repositoryID...)
	)
	if err != nil {
		runtimeinfo.LogError("READ REPO. COMPLETED WITH ERROR: ", err)
		result <- nil
	} else {
		result <- issues
	}
	return
}

func (implement *ImplementRules) getIssueOnlyFromGroupRepositories(result chan []dataModel.IssueModel, repositoryID ...uint) {
	var (
		issues, err = implement.db.GetIssuesOnlyGroupRepositories(repositoryID...)
	)
	if err != nil {
		runtimeinfo.LogError("READ REPO. COMPLETED WITH ERROR: ", err)
		result <- nil
	} else {
		result <- issues
	}
	return
}
