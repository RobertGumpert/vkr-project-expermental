package githubCollectorService

import (
	"fmt"
	"github-gate/app/models/customFieldsModel"
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"strconv"
)

type taskDownloadRepositoryAndRepositoriesContainingKeyword struct {
	service *CollectorService
}

func newTaskDownloadRepositoryAndRepositoriesContainingKeyword(service *CollectorService) *taskDownloadRepositoryAndRepositoriesContainingKeyword {
	return &taskDownloadRepositoryAndRepositoriesContainingKeyword{service: service}
}

func (t *taskDownloadRepositoryAndRepositoriesContainingKeyword) CreateTask(
	taskAppService itask.ITask,
	repositoryDataModel dataModel.RepositoryModel,
	keyWord string,
) (task itask.ITask, err error) {
	if isFilled := t.service.taskManager.QueueIsFilled(33); isFilled {
		return nil, gotasker.ErrorQueueIsFilled
	}
	taskRepositoryDescription, err := t.createTriggerTask(
		taskAppService,
		repositoryDataModel,
	)
	if err != nil {
		return nil, err
	}
	dependentTaskRepositoryIssues, err := t.createDependentTaskIssuesRepository(
		repositoryDataModel,
	)
	if err != nil {
		return nil, err
	}
	dependentTaskRepositoriesKeyword, err := t.createDependentTaskRepositoriesContainingKeyword(
		keyWord,
	)
	if err != nil {
		return nil, err
	}
	dependentsTasks := []itask.ITask{
		dependentTaskRepositoryIssues,
		dependentTaskRepositoriesKeyword,
	}
	t.service.taskManager.SetRunBan(dependentsTasks...)
	return t.service.taskManager.ModifyTaskAsTrigger(
		taskRepositoryDescription,
		dependentsTasks...,
	)
}

func (t *taskDownloadRepositoryAndRepositoriesContainingKeyword) createTriggerTask(
	taskAppService itask.ITask,
	repositoryDataModel dataModel.RepositoryModel,
) (trigger itask.ITask, err error) {
	trigger, err = t.service.createTaskDownloadRepositoriesDescription(
		TaskTypeDownloadCompositeRepositoryAndRepositoriesContainingKeyWord,
		repositoryDataModel,
	)
	if err != nil {
		return nil, err
	}
	trigger.SetEventRunTask(t.service.eventRunTask)
	trigger.GetState().SetEventUpdateState(t.EventUpdateTaskState)
	trigger.GetState().SetCustomFields(&customFieldsModel.Model{
		TaskType: TaskTypeDownloadOnlyDescriptions,
		Fields:   taskAppService,
	})
	return trigger, nil
}

func (t *taskDownloadRepositoryAndRepositoriesContainingKeyword) createDependentTaskIssuesRepository(
	repositoryDataModel dataModel.RepositoryModel,
) (dependentTask itask.ITask, err error) {
	downloadRepositoryIssues, err := t.service.createTaskDownloadRepositoryIssues(
		TaskTypeDownloadCompositeRepositoryAndRepositoriesContainingKeyWord,
		repositoryDataModel,
	)
	if err != nil {
		return nil, err
	}
	downloadRepositoryIssues.SetEventRunTask(t.service.eventRunTask)
	downloadRepositoryIssues.GetState().SetEventUpdateState(t.EventUpdateTaskState)
	downloadRepositoryIssues.GetState().SetCustomFields(&customFieldsModel.Model{
		TaskType: TaskTypeDownloadOnlyIssues,
		Fields:   &repositoryDataModel,
	})
	return downloadRepositoryIssues, nil
}

func (t *taskDownloadRepositoryAndRepositoriesContainingKeyword) createDependentTaskRepositoriesContainingKeyword(
	keyWord string,
) (dependentTask itask.ITask, err error) {
	var (
		trigger itask.ITask
		//
		dependentTasks = make([]itask.ITask, 0)
	)
	trigger, err = t.service.createTaskDownloadRepositoriesByKeyword(
		TaskTypeDownloadCompositeRepositoryAndRepositoriesContainingKeyWord,
		keyWord,
	)
	if err != nil {
		return nil, err
	}
	for next := 0; next < 30; next++ {
		repositoryDataModel := &dataModel.RepositoryModel{
			Name:  keyWord,
			Owner: keyWord,
		}
		dependentTask, err := t.service.createTaskDownloadRepositoryIssues(
			TaskTypeDownloadCompositeRepositoryAndRepositoriesContainingKeyWord,
			*repositoryDataModel,
		)
		if err != nil {
			return nil, err
		}
		//
		dependentTask.SetKey("(download-issues){%s}-" + strconv.Itoa(next))
		dependentTask.SetEventRunTask(t.service.eventRunTask)
		dependentTask.GetState().SetEventUpdateState(t.EventUpdateTaskState)
		dependentTask.GetState().SetCustomFields(&customFieldsModel.Model{
			TaskType: TaskTypeDownloadCompositeByKeyWord,
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
		TaskType: TaskTypeDownloadCompositeByKeyWord,
	})
	//
	return t.service.taskManager.ModifyTaskAsTrigger(
		trigger,
		dependentTasks...,
	)
}

func (t *taskDownloadRepositoryAndRepositoriesContainingKeyword) EventManageTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	customFields := task.GetState().GetCustomFields().(*customFieldsModel.Model)
	switch customFields.GetTaskType() {
	case TaskTypeDownloadCompositeByKeyWord:
		var (
			isTrigger, _                              = task.IsTrigger()
			isDependent, repositoriesByKeyWordTrigger = task.IsDependent()
		)
		if !isTrigger && isDependent {
			isCompletedRepositoriesByKeywordTrigger, issuesRepositoriesByKeywordTasks, err := t.service.taskManager.TriggerIsCompleted(repositoriesByKeyWordTrigger)
			if err != nil {
				// TO DO: error
				return nil
			}
			if isCompletedRepositoriesByKeywordTrigger {
				_, repositoryDescriptionTrigger := repositoriesByKeyWordTrigger.IsDependent()
				isCompletedRepositoryDescriptionTrigger, issuesRepositoryTask, err := t.service.taskManager.TriggerIsCompleted(repositoryDescriptionTrigger)
				if err != nil {
					// TO DO: error
					return nil
				}
				if isCompletedRepositoryDescriptionTrigger {
					deleteTasks = make(map[string]struct{})
					deleteTasks[repositoriesByKeyWordTrigger.GetKey()] = struct{}{}
					deleteTasks[repositoryDescriptionTrigger.GetKey()] = struct{}{}
					for key, _ := range issuesRepositoriesByKeywordTasks {
						deleteTasks[key] = struct{}{}
					}
					for key, _ := range issuesRepositoryTask {
						deleteTasks[key] = struct{}{}
					}
					cf := repositoryDescriptionTrigger.GetState().GetCustomFields().(*customFieldsModel.Model)
					appServiceTask := cf.GetFields().(itask.ITask)
					appServiceTask.GetState().GetCustomFields().(*customFieldsModel.Model).GetFields().(chan itask.ITask) <- appServiceTask
				}
			}
		}
		break
	}
	return deleteTasks
}

func (t *taskDownloadRepositoryAndRepositoriesContainingKeyword) EventUpdateTaskState(task itask.ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool) {
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
					cf := dependentTask.GetState().GetCustomFields().(*customFieldsModel.Model)
					if cf.GetTaskType() == TaskTypeDownloadOnlyIssues {
						dependentTask.GetState().SetRunnable(true)
						dependentTask.GetState().SetCompleted(true)
					}
					if cf.GetTaskType() == TaskTypeDownloadCompositeByKeyWord {
						t.service.taskManager.TakeOffRunBanInQueue(dependentTask)
					}
				}
				cf := task.GetState().GetCustomFields().(*customFieldsModel.Model)
				appServiceTask := cf.GetFields().(itask.ITask)
				appServiceUpdateContext := appServiceTask.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
				appServiceUpdateContext = append(
					appServiceUpdateContext,
					*updateContext,
				)
				appServiceTask.GetState().SetUpdateContext(appServiceUpdateContext)
				task.GetState().SetCompleted(true)
				return nil, false
			}
		} else {
			updateContext = &dataModels[0]
		}
		task.GetState().SetUpdateContext(updateContext)
		//
		if cast.ExecutionTaskStatus.TaskCompleted {
			task.GetState().SetCompleted(true)
			if isTrigger, dependentsTasks := task.IsTrigger(); isTrigger {
				for next := 0; next < len(*dependentsTasks); next++ {
					dependentTask := (*dependentsTasks)[next]
					cf := dependentTask.GetState().GetCustomFields().(*customFieldsModel.Model)
					if cf.GetTaskType() == TaskTypeDownloadOnlyIssues {
						cf := dependentTask.GetState().GetCustomFields().(*customFieldsModel.Model)
						cf.Fields = updateContext
						dependentTask.GetState().SetCustomFields(cf)
						t.service.taskManager.TakeOffRunBanInQueue(dependentTask)
						break
					}
				}
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
		if len(cast.Issues) == 0 {
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
				_, dependentsTasks := trigger.IsTrigger()
				for next := 0; next < len(*dependentsTasks); next++ {
					dependentTask := (*dependentsTasks)[next]
					cf := dependentTask.GetState().GetCustomFields().(*customFieldsModel.Model)
					if cf.GetTaskType() == TaskTypeDownloadCompositeByKeyWord {
						t.service.taskManager.TakeOffRunBanInQueue(dependentTask)
						break
					}
				}
			} else {
				// TO DO: error
				return nil, true
			}

		}
		break
	case TaskTypeDownloadCompositeByKeyWord:
		var (
			isTrigger, _   = task.IsTrigger()
			isDependent, _ = task.IsDependent()
		)
		if isTrigger && isDependent {
			return t.updateTriggerRepositoriesKeyword(task, somethingUpdateContext, customFields)
		}
		if !isTrigger && isDependent {
			return t.updateDependentRepositoriesKeyword(task, somethingUpdateContext, customFields)
		}
		break
	}
	return nil, false
}

func (t *taskDownloadRepositoryAndRepositoriesContainingKeyword) EventRunTask(task itask.ITask) (doTaskAsDefer, sendToErrorChannel bool, err error) {
	return t.service.eventRunTask(task)
}

func (t *taskDownloadRepositoryAndRepositoriesContainingKeyword) updateTriggerRepositoriesKeyword(
	task itask.ITask,
	somethingUpdateContext interface{},
	customFields *customFieldsModel.Model,
) (err error, sendToErrorChannel bool) {
	var(
		dataModels, existRepositories []dataModel.RepositoryModel
	)
	cast := somethingUpdateContext.(*jsonSendFromCollectorRepositoriesByKeyWord)
	if len(cast.Repositories) != 0 {
		dataModels, existRepositories = t.service.writeRepositoriesToDB(cast.Repositories)
		if len(dataModels) == 0 && len(existRepositories) != 0 {
			task.GetState().SetCompleted(true)
			_, repositoryDescriptionTrigger := task.IsDependent()
			cf := repositoryDescriptionTrigger.GetState().GetCustomFields().(*customFieldsModel.Model)
			appServiceTask := cf.GetFields().(itask.ITask)
			appServiceUpdateContext := appServiceTask.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
			appServiceUpdateContext = append(
				appServiceUpdateContext,
				existRepositories...,
			)
			appServiceTask.GetState().SetUpdateContext(appServiceUpdateContext)
			//
			_, dependentsTasks := task.IsTrigger()
			var(
				dependentTask itask.ITask
			)
			if len(*dependentsTasks) != 0 {
				for next := 0; next < len(*dependentsTasks); next++ {
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
		} else {
			updateContext := task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
			updateContext = append(
				updateContext,
				dataModels...,
			)
			task.GetState().SetUpdateContext(updateContext)
		}
	}
	if cast.ExecutionTaskStatus.TaskCompleted {
		task.GetState().SetCompleted(true)
		var (
			deleteDependentTasks []itask.ITask
			//
			deleteTaskKeys = make(map[string]struct{})
			next           = 0
		)
		updateContext := task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		if isTrigger, dependentTasks := task.IsTrigger(); isTrigger {
			for next = 0; next < len(updateContext); next++ {
				dependentTask := (*dependentTasks)[next]
				repositoryDataModel := updateContext[next]
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
			deleteDependentTasks = (*dependentTasks)[next:]
			for _, dependent := range deleteDependentTasks {
				deleteTaskKeys[dependent.GetKey()] = struct{}{}
			}
			t.service.taskManager.DeleteTasksByKeys(deleteTaskKeys)
			*dependentTasks = (*dependentTasks)[:next]
		}
	}
	return nil, false
}

func (t *taskDownloadRepositoryAndRepositoriesContainingKeyword) updateDependentRepositoriesKeyword(
	task itask.ITask,
	somethingUpdateContext interface{},
	customFields *customFieldsModel.Model,
) (err error, sendToErrorChannel bool) {
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
		if isDependentTask, repositoriesByKeywordTrigger := task.IsDependent(); isDependentTask {
			if triggerIsDependent, repositoryDescriptionTrigger := repositoriesByKeywordTrigger.IsDependent(); triggerIsDependent {
				cf := repositoryDescriptionTrigger.GetState().GetCustomFields().(*customFieldsModel.Model)
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
		} else {
			// TO DO: error
			return nil, true
		}
	}
	return nil, false
}
