package appService

import (
	"github-gate/app/models/customFieldsModel"
	"github-gate/app/serivce/githubCollectorService"
	"github-gate/app/serivce/issueIndexerService"
	"github-gate/app/serivce/repositoryIndexerService"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"strings"
)

type taskCompositeNewRepositoryWithExistKeyWord struct {
	taskManager                itask.IManager
	collectorService           *githubCollectorService.CollectorService
	issuesIndexerService       *issueIndexerService.IndexerService
	repositoriesIndexerService *repositoryIndexerService.IndexerService
}

func newTaskCompositeNewRepositoryWithExistKeyWord(
	taskManager itask.IManager,
	collectorService *githubCollectorService.CollectorService,
	issuesIndexerService *issueIndexerService.IndexerService,
	repositoriesIndexerService *repositoryIndexerService.IndexerService,
) *taskCompositeNewRepositoryWithExistKeyWord {
	task := new(taskCompositeNewRepositoryWithExistKeyWord)
	task.taskManager = taskManager
	task.collectorService = collectorService
	task.issuesIndexerService = issuesIndexerService
	task.repositoriesIndexerService = repositoriesIndexerService
	return task
}

func (composite *taskCompositeNewRepositoryWithExistKeyWord) CreateTask(jsonModel *JsonSingleTaskDownloadRepositoriesByName, channel chan itask.ITask) (task itask.ITask, err error) {
	var (
		taskKey = strings.Join([]string{
			"composite-new-repository-with-exist-keyword:{",
			jsonModel.Repositories[0].Name,
			"}",
		}, "")
	)
	downloadTask, err := composite.getTaskForCollector(taskKey, jsonModel, channel)
	if err != nil {
		return nil, err
	}
	issueIndexerTask, err := composite.getTaskForIssueIndexer(taskKey)
	if err != nil {
		return nil, err
	}
	repositoryIndexerTask, err := composite.getTaskForRepositoryIndexer(taskKey)
	if err != nil {
		return nil, err
	}
	composite.taskManager.SetRunBan(issueIndexerTask, repositoryIndexerTask)
	return composite.taskManager.ModifyTaskAsTrigger(downloadTask, repositoryIndexerTask, issueIndexerTask)
}

func (composite *taskCompositeNewRepositoryWithExistKeyWord) getTaskForCollector(taskKey string, jsonModel *JsonSingleTaskDownloadRepositoriesByName, channel chan itask.ITask) (task itask.ITask, err error) {
	var (
		downloadTaskKey = strings.Join([]string{
			taskKey,
			"-[download-repository-by-name]",
		}, "")
		repositoriesNames = make([]string, 0)
		sendContext       = make([]dataModel.RepositoryModel, 0)
		updateContext     = make([]dataModel.RepositoryModel, 0)
		customFields      = &customFieldsModel.Model{
			TaskType: githubCollectorService.TaskTypeDownloadCompositeByName,
			Fields:   channel,
		}
	)
	for _, repository := range jsonModel.Repositories {
		sendContext = append(sendContext, dataModel.RepositoryModel{
			Name:  repository.Name,
			Owner: repository.Owner,
		})
		repositoriesNames = append(repositoriesNames, repository.Name)
	}
	return composite.taskManager.CreateTask(
		CompositeTaskNewRepositoryWithExistWord,
		downloadTaskKey,
		sendContext,
		updateContext,
		customFields,
		composite.EventRunTask,
		composite.EventUpdateTaskState,
	)
}

func (composite *taskCompositeNewRepositoryWithExistKeyWord) getTaskForRepositoryIndexer(taskKey string) (task itask.ITask, err error) {
	var (
		repositoryIndexerTaskKey = strings.Join([]string{
			taskKey,
			"-[repository-indexer-reindexing-for-repository]",
		}, "")
	)
	return composite.taskManager.CreateTask(
		CompositeTaskNewRepositoryWithExistWord,
		repositoryIndexerTaskKey,
		repositoryIndexerService.JsonSendToIndexerReindexingForRepository{
			TaskKey:      repositoryIndexerTaskKey,
			RepositoryID: 0,
		},
		repositoryIndexerService.JsonSendFromIndexerReindexingForRepository{},
		&customFieldsModel.Model{
			TaskType: repositoryIndexerService.TaskTypeReindexingForRepository,
			Fields:   nil,
		},
		composite.EventRunTask,
		composite.EventUpdateTaskState,
	)
}

func (composite *taskCompositeNewRepositoryWithExistKeyWord) getTaskForIssueIndexer(taskKey string) (task itask.ITask, err error) {
	var (
		issueIndexerTaskKey = strings.Join([]string{
			taskKey,
			"-[issue-indexer-compare-group-repositories]",
		}, "")
	)
	return composite.taskManager.CreateTask(
		CompositeTaskNewRepositoryWithExistWord,
		issueIndexerTaskKey,
		issueIndexerService.JsonSendToIndexerCompareGroup{
			TaskKey:                  issueIndexerTaskKey,
			RepositoryID:             0,
			ComparableRepositoriesID: nil,
		},
		issueIndexerService.JsonSendFromIndexerCompareGroup{},
		&customFieldsModel.Model{
			TaskType: issueIndexerService.TaskTypeCompareGroupRepositories,
			Fields:   nil,
		},
		composite.EventRunTask,
		composite.EventUpdateTaskState,
	)
}

func (composite *taskCompositeNewRepositoryWithExistKeyWord) EventManageTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	deleteTasks = make(map[string]struct{})
	taskType := task.GetState().GetCustomFields().(*customFieldsModel.Model).GetTaskType()
	switch taskType {
	case issueIndexerService.TaskTypeCompareGroupRepositories:
		var (
			countCompletedTask int
		)
		if isDependent, trigger := task.IsDependent(); isDependent {
			if !trigger.GetState().IsCompleted() {
				break
			}
			if isTrigger, dependentsTasks := trigger.IsTrigger(); isTrigger {
				for next := 0; next < len(*dependentsTasks); next++ {
					dependentTask := (*dependentsTasks)[next]
					if dependentTask.GetState().IsCompleted() {
						countCompletedTask++
					}
					deleteTasks[dependentTask.GetKey()] = struct{}{}
				}
				if countCompletedTask == len(*dependentsTasks) {
					deleteTasks[trigger.GetKey()] = struct{}{}
				} else {
					deleteTasks = nil
				}
			}
		}
		break
	}
	return deleteTasks
}

func (composite *taskCompositeNewRepositoryWithExistKeyWord) EventRunTask(task itask.ITask) (doTaskAsDefer, sendToErrorChannel bool, err error) {
	taskType := task.GetState().GetCustomFields().(*customFieldsModel.Model).GetTaskType()
	switch taskType {
	case issueIndexerService.TaskTypeCompareGroupRepositories:
		err := composite.issuesIndexerService.CompareGroupRepositories(task)
		if err != nil {
			return true, false, nil
		}
		break
	case repositoryIndexerService.TaskTypeReindexingForRepository:
		err := composite.repositoriesIndexerService.ReindexingForRepository(task)
		if err != nil {
			return true, false, nil
		}
		break
	case githubCollectorService.TaskTypeDownloadCompositeByName:
		err = composite.collectorService.CreateTriggerTaskRepositoriesByName(
			task,
			task.GetState().GetSendContext().([]dataModel.RepositoryModel)...,
		)
		if err != nil {
			return true, false, nil
		}
		break
	}
	return false, false, nil
}

func (composite *taskCompositeNewRepositoryWithExistKeyWord) EventUpdateTaskState(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	taskType := task.GetState().GetCustomFields().(*customFieldsModel.Model).GetTaskType()
	switch taskType {
	case issueIndexerService.TaskTypeCompareGroupRepositories:
		task.GetState().SetCompleted(true)
		break
	case repositoryIndexerService.TaskTypeReindexingForRepository:
		if isDependent, trigger := task.IsDependent(); isDependent {
			if isTrigger, dependentsTasks := trigger.IsTrigger(); isTrigger {
				//
				repository := trigger.GetState().GetUpdateContext().([]dataModel.RepositoryModel)[0]
				group := func() (ids []uint) {
					ids = make([]uint, 0)
					nearest := task.GetState().GetUpdateContext().(repositoryIndexerService.JsonSendFromIndexerReindexingForRepository).Result.NearestRepositoriesID
					for id, _ := range nearest {
						ids = append(ids, id)
					}
					return ids
				}()
				//
				for next := 0; next < len(*dependentsTasks); next++ {
					dependentTask := (*dependentsTasks)[next]
					customFields := dependentTask.GetState().GetCustomFields().(*customFieldsModel.Model)
					if customFields.GetTaskType() == issueIndexerService.TaskTypeCompareGroupRepositories {
						if len(group) == 0 {
							dependentTask.GetState().SetCompleted(true)
							composite.taskManager.SetUpdateForTask(task.GetKey(), nil)
							break
						} else {
							sendContext := dependentTask.GetState().GetSendContext().(issueIndexerService.JsonSendToIndexerCompareGroup)
							sendContext.RepositoryID = repository.ID
							dependentTask.GetState().SetSendContext(sendContext)
							composite.taskManager.TakeOffRunBanInQueue(dependentTask)
							break
						}
					}
				}
			}
			task.GetState().SetCompleted(true)
		}
		break
	case githubCollectorService.TaskTypeDownloadCompositeByName:
		if isTrigger, dependentsTasks := task.IsTrigger(); isTrigger {
			repository := task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)[0]
			for next := 0; next < len(*dependentsTasks); next++ {
				dependentTask := (*dependentsTasks)[next]
				customFields := dependentTask.GetState().GetCustomFields().(*customFieldsModel.Model)
				if customFields.GetTaskType() == repositoryIndexerService.TaskTypeReindexingForRepository {
					sendContext := dependentTask.GetState().GetSendContext().(repositoryIndexerService.JsonSendToIndexerReindexingForRepository)
					sendContext.RepositoryID = repository.ID
					dependentTask.GetState().SetSendContext(sendContext)
					composite.taskManager.TakeOffRunBanInQueue(dependentTask)
					break
				}
			}
		}
		task.GetState().SetCompleted(true)
		break
	}
	return nil, false
}
