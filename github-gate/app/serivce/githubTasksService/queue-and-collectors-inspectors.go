package githubTasksService

import (
	"github-gate/pckg/requests"
	"github-gate/pckg/runtimeinfo"
	"net/http"
)

func (service *GithubTasksService) queueHasFreeSpace(countNewTasks int) bool {
	var queueFreeSpace = true
	if len(service.tasksForCollectorsQueue)+countNewTasks > int(service.config.SizeQueueTasksForGithubCollectors) {
		queueFreeSpace = false
	}
	return queueFreeSpace
}

func (service *GithubTasksService) collectorIsFree(collectorUrl string) bool {
	var collectorFree = true
	getStateUrl := collectorUrl + "/get/state"
	response, err := requests.GET(
		service.client,
		getStateUrl,
		nil,
	)
	if err != nil {
		collectorFree = false
	} else {
		if response.StatusCode != http.StatusOK {
			collectorFree = false
		}
	}
	return collectorFree
}

func (service *GithubTasksService) getFreeCollectors(onlyFirst bool) []string {
	var freeCollectorsAddresses = make([]string, 0)
	for _, collectorUrl := range service.config.GithubCollectorsAddresses {
		getStateUrl := collectorUrl + "/get/state"
		response, err := requests.GET(
			service.client,
			getStateUrl,
			nil,
		)
		if err != nil {
			runtimeinfo.LogError("REQUEST TO COLLECTOR: ", collectorUrl, ", COMPLETED WITH ERROR: ", err)
			continue
		}
		if response.StatusCode == http.StatusOK {
			runtimeinfo.LogInfo("FOUND FREE COLLECTOR: ", collectorUrl)
			freeCollectorsAddresses = append(
				freeCollectorsAddresses,
				collectorUrl,
			)
			if onlyFirst {
				return freeCollectorsAddresses
			}
		}
	}
	if len(freeCollectorsAddresses) == 0 {
		return nil
	}
	return freeCollectorsAddresses
}
