package appService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"log"
)

func (service *AppService) eventManageCompletedTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	deleteTasks = make(map[string]struct{})
	switch task.GetType() {
	case TaskTypeDownloadRepositoryByName:
		repositories := task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		for _, repository := range repositories {
			log.Println(repository.ID)
		}
		deleteTasks[task.GetKey()] = struct{}{}
		break
	case TaskTypeDownloadRepositoryByKeyWord:
		repositories := task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		for _, repository := range repositories {
			log.Println(repository.ID)
		}
		deleteTasks[task.GetKey()] = struct{}{}
		break
	case TaskTypeRepositoryAndRepositoriesByKeyWord:
		repositories := task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		for _, repository := range repositories {
			log.Println(repository.ID)
		}
		deleteTasks[task.GetKey()] = struct{}{}
		break
	case CompositeTaskNewRepositoryWithExistWord:
		deleteTasks = service.taskNewRepositoryWithExistWord.EventManageTasks(task)
		break
	}
	return deleteTasks
}


func (service *AppService) scanChannelForCollectorService() {
	for task := range service.channelResultsFromCollectorService {
		repositories := task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		service.taskManager.SetUpdateForTask(
			task.GetKey(),
			repositories,
		)
	}
}

func (service *AppService) scanChannelForIssueIndexerService() {
	for task := range service.channelResultsFromIssueIndexerService {
		service.taskManager.SetUpdateForTask(
			task.GetKey(),
			task.GetState().GetUpdateContext(),
		)
	}
}

func (service *AppService) scanChannelForRepositoryIndexerService() {
	for task := range service.channelResultsFromRepositoryIndexerService {
		service.taskManager.SetUpdateForTask(
			task.GetKey(),
			task.GetState().GetUpdateContext(),
		)
	}
}