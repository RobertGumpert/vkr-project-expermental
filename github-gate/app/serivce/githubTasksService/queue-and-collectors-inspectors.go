package githubTasksService

import (
	"github-gate/pckg/requests"
	"github-gate/pckg/runtimeinfo"
	"net/http"
)

func (service *GithubTasksService) queueIsBusy(countNewTasks int) bool {
	if service.tasksForCollectors.Count()+countNewTasks > int(service.config.CountTask) {
		return false
	}
	return true
}

func (service *GithubTasksService) getFreeCollectors() []string {
	var freeCollectorsAddresses = make([]string, 0)
	for _, url := range service.config.GithubCollectorsAddresses {
		getStateUrl := url + "/get/state"
		response, err := requests.GET(
			service.client,
			getStateUrl,
			nil,
		)
		if err != nil {
			runtimeinfo.LogError("REQUEST TO COLLECTOR: ", url, ", COMPLETED WITH ERROR: ", err)
			continue
		}
		if response.StatusCode == http.StatusOK {
			runtimeinfo.LogInfo("FOUND FREE COLLECTOR: ", url)
			freeCollectorsAddresses = append(
				freeCollectorsAddresses,
				url,
			)
		}
	}
	if len(freeCollectorsAddresses) == 0 {
		return nil
	}
	return freeCollectorsAddresses
}
