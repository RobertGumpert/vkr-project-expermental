package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/requests"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"net/http"
	"strings"
)

func (service *CollectorService) eventRunTask(task itask.ITask) (err error) {
	var (
		listFreeCollectors   []string
		freeCollectorAddress string
		sendContext          *contextTaskSend
	)
	sendContext = task.GetState().GetSendContext().(*contextTaskSend)
	if sendContext.JSONBody == nil {
		runtimeinfo.LogError(ErrorNotFullSendContext)
		task.GetState().SetError(ErrorNotFullSendContext)
		return ErrorNotFullSendContext
	}
	if sendContext.CollectorEndpoint == "" {
		runtimeinfo.LogError(ErrorNotFullSendContext)
		task.GetState().SetError(ErrorNotFullSendContext)
		return ErrorNotFullSendContext
	}
	listFreeCollectors = service.getFreeCollectors(true)
	if len(listFreeCollectors) == 0 {
		return ErrorNoFreeCollector
	} else {
		freeCollectorAddress = listFreeCollectors[0]
	}
	sendContext.CollectorAddress = freeCollectorAddress
	sendContext.CollectorURL = strings.Join(
		[]string{
			sendContext.CollectorAddress,
			sendContext.CollectorEndpoint,
		},
		"/",
	)
	response, err := requests.POST(service.client, sendContext.CollectorURL, nil, sendContext.JSONBody)
	if err != nil {
		runtimeinfo.LogError(err)
		return ErrorCollectorIsBusy
	}
	if response.StatusCode != http.StatusOK {
		runtimeinfo.LogError(ErrorCollectorIsBusy)
		return ErrorCollectorIsBusy
	}
	return nil
}
