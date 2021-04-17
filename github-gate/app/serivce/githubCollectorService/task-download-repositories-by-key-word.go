package githubCollectorService

import (
	"fmt"
	"github-gate/app/models/customFieldsModel"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
)

type taskDownloadRepositoriesByKeyword struct {
	service *CollectorService
}

func newTaskDownloadRepositoriesByKeyword(service *CollectorService) *taskDownloadRepositoriesByKeyword {
	return &taskDownloadRepositoriesByKeyword{service: service}
}

func (t *taskDownloadRepositoriesByKeyword) CreateTask(
	taskAppService itask.ITask,
	keyWord string,
) (trigger itask.ITask, err error) {
	if t.service.taskManager.QueueIsFilled(31) {
		return nil, ErrorQueueIsFilled
	}
	trigger, err = t.service.createTaskDownloadRepositoriesByKeyword(
		TaskTypeDownloadCompositeByKeyWord,
		keyWord,
	)
	if err != nil {
		return nil, err
	}
	dependentTasks := make([]itask.ITask, 0)
	for next := 0; next < 30; next++ {
		repositoryDataModel := &dataModel.RepositoryModel{
			Name:  keyWord,
			Owner: keyWord,
		}
		dependentTask, err := t.service.createTaskDownloadRepositoryIssues(
			TaskTypeDownloadCompositeByKeyWord,
			*repositoryDataModel,
		)
		if err != nil {
			return nil, err
		}
		//
		dependentTask.SetKey("(download-issues){%s}")
		dependentTask.SetEventRunTask(t.service.eventRunTask)
		dependentTask.GetState().SetEventUpdateState(t.EventUpdateTaskState)
		dependentTask.GetState().SetCustomFields(&customFieldsModel.Model{
			TaskType: TaskTypeDownloadOnlyIssues,
			Fields:   repositoryDataModel,
		})
		//
		dependentTasks = append(
			dependentTasks,
			dependentTask,
		)
	}
	//
	trigger.SetEventRunTask(t.service.eventRunTask)
	trigger.GetState().SetEventUpdateState(t.EventUpdateTaskState)
	trigger.GetState().SetCustomFields(&customFieldsModel.Model{
		TaskType: TaskTypeDownloadOnlyDescriptions,
		Fields:   taskAppService,
	})
	//
	trigger, err = t.service.taskManager.ModifyTaskAsTrigger(
		trigger,
		dependentTasks...,
	)
	return trigger, nil
}

func (t *taskDownloadRepositoriesByKeyword) EventManageTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	customFields := task.GetState().GetCustomFields().(*customFieldsModel.Model)
	switch customFields.GetTaskType() {
	case TaskTypeDownloadOnlyIssues:
		deleteTasks = make(map[string]struct{})
		if isDependent, trigger := task.IsDependent(); isDependent {
			isCompleted, dependentTasks, err := t.service.taskManager.TriggerIsCompleted(trigger)
			if err != nil {
				// TO DO: error
				return nil
			}
			if isCompleted {
				for dependentTaskKey, _ := range dependentTasks {
					deleteTasks[dependentTaskKey] = struct{}{}
				}
				deleteTasks[trigger.GetKey()] = struct{}{}
				cf := trigger.GetState().GetCustomFields().(*customFieldsModel.Model)
				appServiceTask := cf.GetFields().(itask.ITask)
				appServiceTask.GetState().GetCustomFields().(*customFieldsModel.Model).GetFields().(chan itask.ITask) <- appServiceTask
			}
		}
		break
	}
	return deleteTasks
}

func (t *taskDownloadRepositoriesByKeyword) EventUpdateTaskState(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
	customFields := task.GetState().GetCustomFields().(*customFieldsModel.Model)
	switch customFields.GetTaskType() {
	case TaskTypeDownloadOnlyDescriptions:
		var (
			dataModels []dataModel.RepositoryModel
		)
		cast := somethingUpdateContext.(*jsonSendFromCollectorRepositoriesByKeyWord)
		cf := task.GetState().GetCustomFields().(*customFieldsModel.Model)
		appServiceTask := cf.GetFields().(itask.ITask)
		if len(cast.Repositories) != 0 {
			dataModels, _ = t.service.writeRepositoriesToDB(cast.Repositories)
			appServiceUpdateContext := appServiceTask.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
			appServiceUpdateContext = append(
				appServiceUpdateContext,
				dataModels...,
			)
			appServiceTask.GetState().SetUpdateContext(appServiceUpdateContext)
		}
		if cast.ExecutionTaskStatus.TaskCompleted {
			task.GetState().SetCompleted(true)
			var (
				deleteDependentTasks    []itask.ITask
				appServiceUpdateContext []dataModel.RepositoryModel
				//
				deleteTaskKeys = make(map[string]struct{})
				next           = 0
			)
			appServiceUpdateContext = appServiceTask.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
			if isTrigger, dependentsTasks := task.IsTrigger(); isTrigger {
				if len(appServiceUpdateContext) == 0 {
					var(
						dependentTask itask.ITask
					)
					if len(*dependentsTasks) != 0 {
						for next = 0; next < len(*dependentsTasks); next++ {
							dependentTask = (*dependentsTasks)[next]
							dependentTask.GetState().SetRunnable(true)
							dependentTask.GetState().SetCompleted(true)
						}
						if dependentTask != nil {
							t.service.taskManager.SetUpdateForTask(dependentTask.GetKey(), nil)
							return nil, false
						} else {
							// TO DO: error
							return nil, true
						}
					} else {
						// TO DO: error
						return nil, true
					}
				}
				for next = 0; next < len(appServiceUpdateContext); next++ {
					dependentTask := (*dependentsTasks)[next]
					repositoryDataModel := appServiceUpdateContext[next]
					cf := dependentTask.GetState().GetCustomFields().(*customFieldsModel.Model)
					cf.Fields = &repositoryDataModel
					dependentTask.SetKey(fmt.Sprintf(dependentTask.GetKey(), repositoryDataModel.Name))
					sendContext := dependentTask.GetState().GetSendContext().(*contextTaskSend)
					sendContext.JSONBody.(*jsonSendToCollectorRepositoryIssues).TaskKey = dependentTask.GetKey()
					sendContext.JSONBody.(*jsonSendToCollectorRepositoryIssues).Repository.Name = repositoryDataModel.Name
					sendContext.JSONBody.(*jsonSendToCollectorRepositoryIssues).Repository.Owner = repositoryDataModel.Owner
					dependentTask.GetState().SetSendContext(sendContext)
					dependentTask.GetState().SetCustomFields(cf)
				}
				deleteDependentTasks = (*dependentsTasks)[next:]
				for _, dependent := range deleteDependentTasks {
					deleteTaskKeys[dependent.GetKey()] = struct{}{}
				}
				t.service.taskManager.DeleteTasksByKeys(deleteTaskKeys)
				*dependentsTasks = (*dependentsTasks)[:next]
			}
		}
		break
	case TaskTypeDownloadOnlyIssues:
		if task.GetState().IsCompleted() && task.GetState().IsRunnable() {
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
		}
		break
	}
	return nil, false
}

func (t *taskDownloadRepositoriesByKeyword) EventRunTask(task itask.ITask) (doTaskAsDefer, sendToErrorChannel bool, err error) {
	return t.service.eventRunTask(task)
}
