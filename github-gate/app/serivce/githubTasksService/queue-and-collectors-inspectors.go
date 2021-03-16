package githubTasksService

import (
	"github-gate/pckg/requests"
	"github-gate/pckg/runtimeinfo"
	"net/http"
)

func (service *GithubTasksService) queueIsBusy(countNewTasks int) bool {
	if len(service.tasksForCollectors)+countNewTasks > int(service.config.CountTask) {
		return false
	}
	return true
}

func (service *GithubTasksService) collectorIsFree(collectorUrl string) bool {
	getStateUrl := collectorUrl + "/get/state"
	response, err := requests.GET(
		service.client,
		getStateUrl,
		nil,
	)
	if err != nil {
		return false
	}
	if response.StatusCode != http.StatusOK {
		return false
	}
	return true
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
