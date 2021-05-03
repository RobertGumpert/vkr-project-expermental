package appService

import (
	"github-gate/app/models/customFieldsModel"
	"github-gate/app/serivce/issueIndexerService"
	"github-gate/app/serivce/repositoryIndexerService"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"strconv"
	"strings"
)

type taskExistRepository struct {
	appService *AppService
}

func newTaskExistRepository(appService *AppService) *taskExistRepository {
	return &taskExistRepository{appService: appService}
}

func (t *taskExistRepository) CreateTask(jsonModel *JsonExistRepository) (task itask.ITask, err error) {
	var (
		uniqueKey, taskKey string
		repositoryId       uint
	)
	if jsonModel.RepositoryID != 0 {
		_, err := t.appService.db.GetRepositoryByID(jsonModel.RepositoryID)
		if err != nil {
			return nil, ErrorEmptyOrIncompleteJSONData
		}
		repositoryId = jsonModel.RepositoryID
		uniqueKey = strconv.Itoa(int(jsonModel.RepositoryID))
	} else {
		if strings.TrimSpace(jsonModel.Repository.Name) != "" {
			m, err := t.appService.db.GetRepositoryByName(jsonModel.Repository.Name)
			if err != nil {
				return nil, ErrorEmptyOrIncompleteJSONData
			}
			repositoryId = m.ID
			uniqueKey = jsonModel.Repository.Name
		} else {
			return nil, ErrorEmptyOrIncompleteJSONData
		}
	}
	taskKey = strings.Join([]string{
		"exist-repository:{",
		uniqueKey,
		"}",
	}, "")
	issueIndexerTask, err := t.getTaskForIssueIndexer(taskKey)
	if err != nil {
		return nil, err
	}
	repositoryIndexerTask, err := t.getTaskForRepositoryIndexer(taskKey, repositoryId, jsonModel)
	if err != nil {
		return nil, err
	}
	return t.appService.taskManager.ModifyTaskAsTrigger(repositoryIndexerTask, issueIndexerTask)
}

func (t *taskExistRepository) getTaskForRepositoryIndexer(taskKey string, repositoryId uint, jsonModel *JsonExistRepository) (task itask.ITask, err error) {
	var (
		repositoryIndexerTaskKey = strings.Join([]string{
			taskKey,
			"-[repository-indexer-reindexing-for-repository]",
		}, "")
		sendContext = &repositoryIndexerService.JsonSendToIndexerReindexingForRepository{
			TaskKey:      repositoryIndexerTaskKey,
			RepositoryID: repositoryId,
		}
		updateContext = &repositoryIndexerService.JsonSendFromIndexerReindexingForRepository{}
		customFields  = &customFieldsModel.Model{
			TaskType: repositoryIndexerService.TaskTypeReindexingForRepository,
			Fields:   nil,
			Context:  jsonModel.UserRequest,
		}
	)
	return t.appService.taskManager.CreateTask(
		TaskTypeExistRepository,
		repositoryIndexerTaskKey,
		sendContext,
		updateContext,
		customFields,
		t.EventRunTask,
		t.EventUpdateTaskState,
	)
}

func (t *taskExistRepository) getTaskForIssueIndexer(taskKey string) (task itask.ITask, err error) {
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
		TaskTypeExistRepository,
		issueIndexerTaskKey,
		sendContext,
		updateContext,
		customFields,
		t.EventRunTask,
		t.EventUpdateTaskState,
	)
}

func (t *taskExistRepository) EventManageTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	taskType := task.GetState().GetCustomFields().(*customFieldsModel.Model).GetTaskType()
	switch taskType {
	case issueIndexerService.TaskTypeCompareIssuesGroupRepositories:
		deleteTasks = make(map[string]struct{})
		_, trigger := task.IsDependent()
		_, dependentsTasks := trigger.IsTrigger()
		countCompletedTask := 0
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
		break
	}
	return deleteTasks
}

func (t *taskExistRepository) EventUpdateTaskState(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	taskType := task.GetState().GetCustomFields().(*customFieldsModel.Model).GetTaskType()
	switch taskType {
	case issueIndexerService.TaskTypeCompareIssuesGroupRepositories:
		task.GetState().SetCompleted(true)
		break
	case repositoryIndexerService.TaskTypeReindexingForRepository:
		var (
			_, dependentsTasks = task.IsTrigger()
			intersections      []dataModel.NumberIssueIntersectionsModel
			mpIds              map[uint]struct{}
			//
			intersectionsIdMap = func() {
				mpIds = make(map[uint]struct{})
				for _, inter := range intersections {
					mpIds[inter.ComparableRepositoryID] = struct{}{}
				}
			}
			repositoryId  = task.GetState().GetSendContext().(*repositoryIndexerService.JsonSendToIndexerReindexingForRepository).RepositoryID
			nearest       = task.GetState().GetUpdateContext().(*repositoryIndexerService.JsonSendFromIndexerReindexingForRepository).Result.NearestRepositoriesID
			group         = make([]uint, 0)
		)
		intersections, _ = t.appService.db.GetNumberIntersectionsForRepository(repositoryId)
		intersectionsIdMap()
		for id, _ := range nearest {
			if _, exist := mpIds[id]; !exist {
				group = append(group, id)
			}
		}
		for next := 0; next < len(*dependentsTasks); next++ {
			dependentTask := (*dependentsTasks)[next]
			customFields := dependentTask.GetState().GetCustomFields().(*customFieldsModel.Model)
			if customFields.GetTaskType() == issueIndexerService.TaskTypeCompareIssuesGroupRepositories {
				if len(group) == 0 {
					dependentTask.GetState().SetCompleted(true)
					dependentTask.GetState().SetRunnable(true)
					t.appService.taskManager.SetUpdateForTask(dependentTask.GetKey(), nil)
					break
				} else {
					sendContext := dependentTask.GetState().GetSendContext().(*issueIndexerService.JsonSendToIndexerCompareGroup)
					sendContext.RepositoryID = repositoryId
					sendContext.ComparableRepositoriesID = group
					dependentTask.GetState().SetSendContext(sendContext)
					break
				}
			}
		}
		task.GetState().SetCompleted(true)
		break
	}
	return nil, false
}

func (t *taskExistRepository) EventRunTask(task itask.ITask) (doTaskAsDefer, sendToErrorChannel bool, err error) {
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
	}
	return false, false, nil
}
