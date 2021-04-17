package githubCollectorService

import (
	"github-gate/app/models/customFieldsModel"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"strings"
)

type taskDownloadRepositoriesByName struct {
	service *CollectorService
}

func newTaskDownloadRepositoriesByName(service *CollectorService) *taskDownloadRepositoriesByName {
	return &taskDownloadRepositoriesByName{service: service}
}

func (t *taskDownloadRepositoriesByName) CreateTask(
	taskAppService itask.ITask,
	repositoryDataModels ...dataModel.RepositoryModel,
) (triggers []itask.ITask, err error) {
	if isFilled := t.service.taskManager.QueueIsFilled(int64(len(repositoryDataModels) * 2)); isFilled {
		return nil, ErrorQueueIsFilled
	}
	triggers = make([]itask.ITask, 0)
	for _, repositoryDataModel := range repositoryDataModels {
		if strings.TrimSpace(repositoryDataModel.Name) == "" ||
			strings.TrimSpace(repositoryDataModel.Owner) == "" {
			return nil, ErrorNoneCorrectData
		}
		_, err := t.service.repository.GetRepositoryByName(
			repositoryDataModel.Name,
		)
		if err == nil {
			return nil, ErrorRepositoryIsExist
		}
		taskRepositoryDescriptions, err := t.service.createTaskDownloadRepositoriesDescription(
			TaskTypeDownloadCompositeByName,
			repositoryDataModel,
		)
		if err != nil {
			return nil, err
		}
		taskRepositoryIssues, err := t.service.createTaskDownloadRepositoryIssues(
			TaskTypeDownloadCompositeByName,
			repositoryDataModel,
		)
		if err != nil {
			return nil, err
		}
		//
		taskRepositoryDescriptions.SetEventRunTask(t.service.eventRunTask)
		taskRepositoryDescriptions.GetState().SetEventUpdateState(t.EventUpdateTaskState)
		taskRepositoryDescriptions.GetState().SetCustomFields(&customFieldsModel.Model{
			TaskType: TaskTypeDownloadOnlyDescriptions,
			Fields:   taskAppService,
		})
		//
		taskRepositoryIssues.SetEventRunTask(t.service.eventRunTask)
		taskRepositoryIssues.GetState().SetEventUpdateState(t.EventUpdateTaskState)
		taskRepositoryIssues.GetState().SetCustomFields(&customFieldsModel.Model{
			TaskType: TaskTypeDownloadOnlyIssues,
			Fields:   &repositoryDataModel,
		})
		//
		trigger, err := t.service.taskManager.ModifyTaskAsTrigger(
			taskRepositoryDescriptions,
			taskRepositoryIssues,
		)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, trigger)
	}
	return triggers, nil
}

func (t *taskDownloadRepositoriesByName) EventManageTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	customFields := task.GetState().GetCustomFields().(*customFieldsModel.Model)
	switch customFields.GetTaskType() {
	case TaskTypeDownloadOnlyIssues:
		deleteTasks = make(map[string]struct{})
		if isDependent, trigger := task.IsDependent(); isDependent {
			cf := trigger.GetState().GetCustomFields().(*customFieldsModel.Model)
			appServiceTask := cf.GetFields().(itask.ITask)
			appServiceUpdateContext := appServiceTask.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
			appServiceSendContext := appServiceTask.GetState().GetSendContext().([]dataModel.RepositoryModel)
			if len(appServiceUpdateContext) == len(appServiceSendContext) {
				appServiceTask.GetState().GetCustomFields().(*customFieldsModel.Model).GetFields().(chan itask.ITask) <- appServiceTask
			}
			deleteTasks[task.GetKey()] = struct{}{}
			deleteTasks[trigger.GetKey()] = struct{}{}
		} else {
			// TO DO: error
			return deleteTasks
		}
		break
	}
	return deleteTasks
}

func (t *taskDownloadRepositoriesByName) EventUpdateTaskState(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	customFields := task.GetState().GetCustomFields().(*customFieldsModel.Model)
	switch customFields.GetTaskType() {
	case TaskTypeDownloadOnlyDescriptions:
		var(
			updateContext *dataModel.RepositoryModel
		)
		cast := somethingUpdateContext.(*jsonSendFromCollectorDescriptionsRepositories)
		if len(cast.Repositories) == 0 {
			// TO DO: error
			return nil, true
		}
		dataModels, existRepositories := t.service.writeRepositoriesToDB(cast.Repositories)
		if len(existRepositories) != 0 {
			updateContext = &existRepositories[0]
			list, _ := t.service.repository.GetIssueRepository(updateContext.ID)
			if len(list) != 0 {
				_, dependentsTasks := task.IsTrigger()
				var(
					dependentTask itask.ITask
				)
				for next := 0; next < len(*dependentsTasks); next++ {
					dependentTask = (*dependentsTasks)[next]
					dependentTask.GetState().SetRunnable(true)
					dependentTask.GetState().SetCompleted(true)
				}
				cf := task.GetState().GetCustomFields().(*customFieldsModel.Model)
				appServiceTask := cf.GetFields().(itask.ITask)
				appServiceUpdateContext := appServiceTask.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
				appServiceUpdateContext = append(
					appServiceUpdateContext,
					*updateContext,
				)
				appServiceTask.GetState().SetUpdateContext(appServiceUpdateContext)
				t.service.taskManager.SetUpdateForTask(task.GetKey(), nil)
				return nil, false
			}
		} else {
			updateContext = &dataModels[0]
		}
		task.GetState().SetUpdateContext(updateContext)
		//
		if cast.ExecutionTaskStatus.TaskCompleted {
			task.GetState().SetCompleted(true)
			if isTrigger, dependentTasks := task.IsTrigger(); isTrigger {
				dependentTask := (*dependentTasks)[0]
				cf := dependentTask.GetState().GetCustomFields().(*customFieldsModel.Model)
				cf.Fields = updateContext
				dependentTask.GetState().SetCustomFields(cf)
			} else {
				// TO DO: error
				return nil, true
			}
		}
		break
	case TaskTypeDownloadOnlyIssues:
		if task.GetState().IsRunnable() && task.GetState().IsCompleted() {
			return nil, false
		}
		cast := somethingUpdateContext.(*jsonSendFromCollectorRepositoryIssues)
		var (
			repositoryId    uint
			repositoryModel *dataModel.RepositoryModel
		)
		repositoryModel = customFields.GetFields().(*dataModel.RepositoryModel)
		if repositoryModel == nil || repositoryModel.ID == 0 {
			// TO DO: error
			return nil, true
		} else {
			repositoryId = repositoryModel.ID
		}
		if len(cast.Issues) != 0 {
			t.service.writeIssuesToDB(cast.Issues, repositoryId)
		}
		if cast.ExecutionTaskStatus.TaskCompleted {
			task.GetState().SetCompleted(true)
			if isDependent, trigger := task.IsDependent(); isDependent {
				cf := trigger.GetState().GetCustomFields().(*customFieldsModel.Model)
				appServiceTask := cf.GetFields().(itask.ITask)
				appServiceUpdateContext := appServiceTask.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
				appServiceUpdateContext = append(
					appServiceUpdateContext,
					*repositoryModel,
				)
				appServiceTask.GetState().SetUpdateContext(appServiceUpdateContext)
			} else {
				// TO DO: error
				return nil, true
			}
		}
		break
	}
	return nil, false
}

func (t *taskDownloadRepositoriesByName) EventRunTask(task itask.ITask) (doTaskAsDefer, sendToErrorChannel bool, err error) {
	return t.service.eventRunTask(task)
}
