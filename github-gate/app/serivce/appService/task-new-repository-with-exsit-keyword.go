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

type taskNewRepositoryWithExistKeyWord struct {
	appService *AppService
}

func newTaskNewRepositoryWithExistKeyWord(appService *AppService) *taskNewRepositoryWithExistKeyWord {
	task := new(taskNewRepositoryWithExistKeyWord)
	task.appService = appService
	return task
}

func (t *taskNewRepositoryWithExistKeyWord) CreateTask(jsonModel *JsonNewRepositoryWithExistKeyword) (task itask.ITask, err error) {
	var (
		taskKey = strings.Join([]string{
			"new-repository-with-exist-keyword:{",
			jsonModel.Repositories[0].Name,
			"}",
		}, "")
	)
	downloadTask, err := t.getTaskForCollector(taskKey, jsonModel)
	if err != nil {
		return nil, err
	}
	issueIndexerTask, err := t.getTaskForIssueIndexer(taskKey)
	if err != nil {
		return nil, err
	}
	repositoryIndexerTask, err := t.getTaskForRepositoryIndexer(taskKey)
	if err != nil {
		return nil, err
	}
	t.appService.taskManager.SetRunBan(issueIndexerTask, repositoryIndexerTask)
	return t.appService.taskManager.ModifyTaskAsTrigger(downloadTask, repositoryIndexerTask, issueIndexerTask)
}

func (t *taskNewRepositoryWithExistKeyWord) getTaskForCollector(taskKey string, jsonModel *JsonNewRepositoryWithExistKeyword) (task itask.ITask, err error) {
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
			Fields:   t.appService.channelResultsFromCollectorService,
			Context:  jsonModel.UserRequest,
		}
	)
	for _, repository := range jsonModel.Repositories {
		sendContext = append(sendContext, dataModel.RepositoryModel{
			Name:  repository.Name,
			Owner: repository.Owner,
		})
		repositoriesNames = append(repositoriesNames, repository.Name)
	}
	return t.appService.taskManager.CreateTask(
		TaskTypeNewRepositoryWithExistKeyword,
		downloadTaskKey,
		sendContext,
		updateContext,
		customFields,
		t.EventRunTask,
		t.EventUpdateTaskState,
	)
}

func (t *taskNewRepositoryWithExistKeyWord) getTaskForRepositoryIndexer(taskKey string) (task itask.ITask, err error) {
	var (
		repositoryIndexerTaskKey = strings.Join([]string{
			taskKey,
			"-[repository-indexer-reindexing-for-repository]",
		}, "")
		sendContext = &repositoryIndexerService.JsonSendToIndexerReindexingForRepository{
			TaskKey:      repositoryIndexerTaskKey,
			RepositoryID: 0,
		}
		updateContext = &repositoryIndexerService.JsonSendFromIndexerReindexingForRepository{}
		customFields  = &customFieldsModel.Model{
			TaskType: repositoryIndexerService.TaskTypeReindexingForRepository,
			Fields:   nil,
		}
	)
	return t.appService.taskManager.CreateTask(
		TaskTypeNewRepositoryWithExistKeyword,
		repositoryIndexerTaskKey,
		sendContext,
		updateContext,
		customFields,
		t.EventRunTask,
		t.EventUpdateTaskState,
	)
}

func (t *taskNewRepositoryWithExistKeyWord) getTaskForIssueIndexer(taskKey string) (task itask.ITask, err error) {
	var (
		issueIndexerTaskKey = strings.Join([]string{
			taskKey,
			"-[issue-indexer-compare-group-repositories]",
		}, "")
		sendContext = &issueIndexerService.JsonSendToIndexerCompareGroup{
			TaskKey:                  issueIndexerTaskKey,
			RepositoryID:             0,
			ComparableRepositoriesID: nil,
		}
		updateContext = &issueIndexerService.JsonSendFromIndexerCompareGroup{}
		customFields  = &customFieldsModel.Model{
			TaskType: issueIndexerService.TaskTypeCompareIssuesGroupRepositories,
			Fields:   nil,
		}
	)
	return t.appService.taskManager.CreateTask(
		TaskTypeNewRepositoryWithExistKeyword,
		issueIndexerTaskKey,
		sendContext,
		updateContext,
		customFields,
		t.EventRunTask,
		t.EventUpdateTaskState,
	)
}

func (t *taskNewRepositoryWithExistKeyWord) EventManageTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	deleteTasks = make(map[string]struct{})
	taskType := task.GetState().GetCustomFields().(*customFieldsModel.Model).GetTaskType()
	switch taskType {
	case issueIndexerService.TaskTypeCompareIssuesGroupRepositories:
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
					customFields := dependentTask.GetState().GetCustomFields().(*customFieldsModel.Model)
					if customFields.GetTaskType() == repositoryIndexerService.TaskTypeReindexingForRepository {
						nearest := dependentTask.GetState().GetUpdateContext().(*repositoryIndexerService.JsonSendFromIndexerReindexingForRepository)
						trigger.GetState().SetUpdateContext(nearest)
					}
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

func (t *taskNewRepositoryWithExistKeyWord) EventRunTask(task itask.ITask) (doTaskAsDefer, sendToErrorChannel bool, err error) {
	taskType := task.GetState().GetCustomFields().(*customFieldsModel.Model).GetTaskType()
	switch taskType {
	case issueIndexerService.TaskTypeCompareIssuesGroupRepositories:
		err := t.appService.serviceForIssueIndexer.CompareGroupRepositories(task)
		if err != nil {
			return true, false, nil
		}
		break
	case repositoryIndexerService.TaskTypeReindexingForRepository:
		err := t.appService.serviceForRepositoryIndexer.ReindexingForRepository(task)
		if err != nil {
			return true, false, nil
		}
		break
	case githubCollectorService.TaskTypeDownloadCompositeByName:
		err = t.appService.serviceForCollector.CreateTaskRepositoriesByName(
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

func (t *taskNewRepositoryWithExistKeyWord) EventUpdateTaskState(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	taskType := task.GetState().GetCustomFields().(*customFieldsModel.Model).GetTaskType()
	switch taskType {
	case issueIndexerService.TaskTypeCompareIssuesGroupRepositories:
		task.GetState().SetCompleted(true)
		break
	case repositoryIndexerService.TaskTypeReindexingForRepository:
		if isDependent, trigger := task.IsDependent(); isDependent {
			if isTrigger, dependentsTasks := trigger.IsTrigger(); isTrigger {
				//
				repository := trigger.GetState().GetUpdateContext().([]dataModel.RepositoryModel)[0]
				group := func() (ids []uint) {
					ids = make([]uint, 0)
					nearest := task.GetState().GetUpdateContext().(*repositoryIndexerService.JsonSendFromIndexerReindexingForRepository).Result.NearestRepositoriesID
					for id, _ := range nearest {
						ids = append(ids, id)
					}
					return ids
				}()
				//
				for next := 0; next < len(*dependentsTasks); next++ {
					dependentTask := (*dependentsTasks)[next]
					customFields := dependentTask.GetState().GetCustomFields().(*customFieldsModel.Model)
					if customFields.GetTaskType() == issueIndexerService.TaskTypeCompareIssuesGroupRepositories {
						t.appService.taskManager.TakeOffRunBanInQueue(dependentTask)
						if len(group) == 0 {
							dependentTask.GetState().SetCompleted(true)
							dependentTask.GetState().SetRunnable(true)
							t.appService.taskManager.SetUpdateForTask(dependentTask.GetKey(), nil)
							break
						} else {
							sendContext := dependentTask.GetState().GetSendContext().(*issueIndexerService.JsonSendToIndexerCompareGroup)
							sendContext.RepositoryID = repository.ID
							sendContext.ComparableRepositoriesID = group
							dependentTask.GetState().SetSendContext(sendContext)
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
					sendContext := dependentTask.GetState().GetSendContext().(*repositoryIndexerService.JsonSendToIndexerReindexingForRepository)
					sendContext.RepositoryID = repository.ID
					dependentTask.GetState().SetSendContext(sendContext)
					t.appService.taskManager.TakeOffRunBanInQueue(dependentTask)
					break
				}
			}
		}
		task.GetState().SetCompleted(true)
		break
	}
	return nil, false
}
