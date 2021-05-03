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

type taskNewRepositoryWithNewKeyword struct {
	appService *AppService
}

func newTaskNewRepositoryWithNewKeyword(appService *AppService) *taskNewRepositoryWithNewKeyword {
	return &taskNewRepositoryWithNewKeyword{appService: appService}
}

func (t *taskNewRepositoryWithNewKeyword) CreateTask(jsonModel *JsonNewRepositoryWithNewKeyword) (task itask.ITask, err error) {
	var (
		taskKey = strings.Join([]string{
			"new-repository-with-new-keyword:{",
			jsonModel.Repository.Name,
			"}:{",
			jsonModel.Keyword,
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

func (t *taskNewRepositoryWithNewKeyword) getTaskForCollector(taskKey string, jsonModel *JsonNewRepositoryWithNewKeyword) (task itask.ITask, err error) {
	var (
		downloadTaskKey = strings.Join([]string{
			taskKey,
			"-[download-repository-and-repositories-by-keyword]",
		}, "")
		sendContext   = jsonModel
		updateContext = make([]dataModel.RepositoryModel, 0)
		customFields  = &customFieldsModel.Model{
			TaskType: githubCollectorService.TaskTypeDownloadCompositeRepositoryAndRepositoriesContainingKeyWord,
			Fields:   t.appService.channelResultsFromCollectorService,
			Context:  jsonModel.UserRequest,
		}
	)
	return t.appService.taskManager.CreateTask(
		TaskTypeNewRepositoryWithNewKeyword,
		downloadTaskKey,
		sendContext,
		updateContext,
		customFields,
		t.EventRunTask,
		t.EventUpdateTaskState,
	)
}

func (t *taskNewRepositoryWithNewKeyword) getTaskForRepositoryIndexer(taskKey string) (task itask.ITask, err error) {
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
		TaskTypeNewRepositoryWithNewKeyword,
		repositoryIndexerTaskKey,
		sendContext,
		updateContext,
		customFields,
		t.EventRunTask,
		t.EventUpdateTaskState,
	)
}

func (t *taskNewRepositoryWithNewKeyword) getTaskForIssueIndexer(taskKey string) (task itask.ITask, err error) {
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
		TaskTypeNewRepositoryWithNewKeyword,
		issueIndexerTaskKey,
		sendContext,
		updateContext,
		customFields,
		t.EventRunTask,
		t.EventUpdateTaskState,
	)
}

func (t *taskNewRepositoryWithNewKeyword) EventManageTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
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

func (t *taskNewRepositoryWithNewKeyword) EventUpdateTaskState(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	taskType := task.GetState().GetCustomFields().(*customFieldsModel.Model).GetTaskType()
	switch taskType {
	case githubCollectorService.TaskTypeDownloadCompositeRepositoryAndRepositoriesContainingKeyWord:
		if isTrigger, dependentsTasks := task.IsTrigger(); isTrigger {
			var(
				repositoryDataModel dataModel.RepositoryModel
				//
				repositoryJsonModel = task.GetState().GetSendContext().(*JsonNewRepositoryWithNewKeyword).Repository
				updateContext = task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
			)
			for next := 0; next < len(updateContext); next++ {
				downloadRepositoryDataModel := updateContext[next]
				if downloadRepositoryDataModel.Name == repositoryJsonModel.Name &&
					downloadRepositoryDataModel.Owner == repositoryJsonModel.Owner {
					repositoryDataModel = downloadRepositoryDataModel
					break
				}
			}
			for next := 0; next < len(*dependentsTasks); next++ {
				dependentTask := (*dependentsTasks)[next]
				customFields := dependentTask.GetState().GetCustomFields().(*customFieldsModel.Model)
				if customFields.GetTaskType() == repositoryIndexerService.TaskTypeReindexingForRepository {
					sendContext := dependentTask.GetState().GetSendContext().(*repositoryIndexerService.JsonSendToIndexerReindexingForRepository)
					sendContext.RepositoryID = repositoryDataModel.ID
					dependentTask.GetState().SetSendContext(sendContext)
					t.appService.taskManager.TakeOffRunBanInQueue(dependentTask)
					break
				}
			}
		}
		task.GetState().SetCompleted(true)
		break
	case repositoryIndexerService.TaskTypeReindexingForRepository:
		if isDependent, trigger := task.IsDependent(); isDependent {
			if isTrigger, dependentsTasks := trigger.IsTrigger(); isTrigger {
				var(
					repositoryDataModel dataModel.RepositoryModel
					//
					group =  make([]uint, 0)
					repositoryJsonModel = trigger.GetState().GetSendContext().(*JsonNewRepositoryWithNewKeyword).Repository
					updateContext = trigger.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
					nearest = task.GetState().GetUpdateContext().(*repositoryIndexerService.JsonSendFromIndexerReindexingForRepository).Result.NearestRepositoriesID
				)
				for next := 0; next < len(updateContext); next++ {
					downloadRepositoryDataModel := updateContext[next]
					if downloadRepositoryDataModel.Name == repositoryJsonModel.Name &&
						downloadRepositoryDataModel.Owner == repositoryJsonModel.Owner {
						repositoryDataModel = downloadRepositoryDataModel
						break
					}
				}
				for id, _ := range nearest {
					group = append(group, id)
				}
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
							sendContext.RepositoryID = repositoryDataModel.ID
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
	case issueIndexerService.TaskTypeCompareIssuesGroupRepositories:
		task.GetState().SetCompleted(true)
		break
	}
	return nil, false
}

func (t *taskNewRepositoryWithNewKeyword) EventRunTask(task itask.ITask) (doTaskAsDefer, sendToErrorChannel bool, err error) {
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
	case githubCollectorService.TaskTypeDownloadCompositeRepositoryAndRepositoriesContainingKeyWord:
		var (
			sendContext         = task.GetState().GetSendContext().(*JsonNewRepositoryWithNewKeyword)
			repositoryDataModel = dataModel.RepositoryModel{
				Name:  sendContext.Repository.Name,
				Owner: sendContext.Repository.Owner,
			}
		)
		err = t.appService.serviceForCollector.CreateTaskRepositoryAndRepositoriesContainingKeyword(
			task,
			repositoryDataModel,
			sendContext.Keyword,
		)
		if err != nil {
			return true, false, nil
		}
		break
	}
	return false, false, nil
}
