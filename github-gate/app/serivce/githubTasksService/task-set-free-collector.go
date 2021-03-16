package githubTasksService

import (
	"fmt"
)

func (service *GithubTasksService) findAndSetCollectorForNewTask(taskForCollector *TaskForCollector, collectorEndpoint string) bool {
	var (
		isSetCollector          = false
		freeCollectorsAddresses = service.getFreeCollectors(true)
	)
	if freeCollectorsAddresses != nil {
		freeCollectorAddress := freeCollectorsAddresses[0]
		taskForCollector.taskDetails.SetCollectorAddress(freeCollectorAddress)
		taskForCollector.taskDetails.SetCollectorEndpoint(collectorEndpoint)
		taskForCollector.taskDetails.SetCollectorURL(
			fmt.Sprintf(
				"%s/%s",
				freeCollectorAddress,
				collectorEndpoint,
			),
		)
		isSetCollector = true
	}
	return isSetCollector
}

func (service *GithubTasksService) setNewCollectorForTask(taskForCollector *TaskForCollector, newCollectorAddress string) {
	taskForCollector.taskDetails.collectorAddress = newCollectorAddress
	taskForCollector.taskDetails.collectorURL = fmt.Sprintf(
		"%s/%s",
		newCollectorAddress,
		taskForCollector.taskDetails.collectorEndpoint,
	)
}
