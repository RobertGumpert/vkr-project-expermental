package githubCollectorService

import (
	"github.com/RobertGumpert/vkr-pckg/requests"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"net/http"
)

func (service *CollectorService) collectorIsFree(collectorUrl string) bool {
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

func (service *CollectorService) getFreeCollectors(onlyFirst bool) []string {
	var freeCollectorsAddresses = make([]string, 0)
	for _, collectorUrl := range service.config.GithubCollectorsAddresses {
		getStateUrl := collectorUrl + "/api/task/get/state"
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