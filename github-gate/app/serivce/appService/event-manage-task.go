package appService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/dataModel"
	"log"
)

func (service *AppService) eventManageCompletedTasks(task itask.ITask) (deleteTasks map[string]struct{}) {
	deleteTasks = make(map[string]struct{})
	switch task.GetType() {
	case SingleTaskDownloadRepositoryByName:
		repositories := task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		for _, repository := range repositories {
			log.Println(repository.ID)
		}
		deleteTasks[task.GetKey()] = struct{}{}
		break
	case SingleTaskDownloadRepositoryByKeyWord:
		repositories := task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		for _, repository := range repositories {
			log.Println(repository.ID)
		}
		deleteTasks[task.GetKey()] = struct{}{}
		break
	case SingleTaskRepositoryAndRepositoriesByKeyWord:
		repositories := task.GetState().GetUpdateContext().([]dataModel.RepositoryModel)
		for _, repository := range repositories {
			log.Println(repository.ID)
		}
		deleteTasks[task.GetKey()] = struct{}{}
		break
	}
	return deleteTasks
}
