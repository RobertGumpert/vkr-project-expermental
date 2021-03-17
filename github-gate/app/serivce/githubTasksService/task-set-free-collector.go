package githubTasksService

import (
	"fmt"
)

func (service *GithubTasksService) findAndSetCollectorForNewTask(taskForCollector *TaskForCollector, collectorEndpoint string) bool {
	var (
		nonFreeCollectors       = false
		freeCollectorsAddresses = service.getFreeCollectors(true)
	)
	if freeCollectorsAddresses != nil {
		freeCollectorAddress := freeCollectorsAddresses[0]
		taskForCollector.details.SetCollectorAddress(freeCollectorAddress)
		taskForCollector.details.SetCollectorEndpoint(collectorEndpoint)
		taskForCollector.details.SetCollectorURL(
			fmt.Sprintf(
				"%s/%s",
				taskForCollector.details.GetCollectorAddress(),
				taskForCollector.details.GetCollectorEndpoint(),
			),
		)
	} else {
		nonFreeCollectors = true
	}
	return nonFreeCollectors
}

func (service *GithubTasksService) setNewCollectorForTask(taskForCollector *TaskForCollector, newCollectorAddress string) {
	taskForCollector.details.SetCollectorAddress(newCollectorAddress)
	taskForCollector.details.SetCollectorURL(
		fmt.Sprintf(
			"%s/%s",
			taskForCollector.details.GetCollectorAddress(),
			taskForCollector.details.GetCollectorEndpoint(),
		),
	)
}
