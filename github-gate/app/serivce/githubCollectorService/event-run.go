package githubCollectorService

import (
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/vkr-pckg/requests"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"net/http"
	"strings"
)

func (service *CollectorService) eventRunTask(task itask.ITask) (doTaskAsDefer, sendToErrorChannel bool, err error) {
	var (
		listFreeCollectors   []string
		freeCollectorAddress string
		sendContext          *contextTaskSend
	)
	sendContext = task.GetState().GetSendContext().(*contextTaskSend)
	if sendContext.JSONBody == nil {
		runtimeinfo.LogError(ErrorNoneCorrectData)
		task.GetState().SetError(ErrorNoneCorrectData)
		return true, false, nil
	}
	if sendContext.CollectorEndpoint == "" {
		runtimeinfo.LogError(ErrorNoneCorrectData)
		task.GetState().SetError(ErrorNoneCorrectData)
		return true, false, nil
	}
	listFreeCollectors = service.getFreeCollectors(true)
	if len(listFreeCollectors) == 0 {
		return true, false, nil
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
		return true, false, nil
	}
	if response.StatusCode != http.StatusOK {
		runtimeinfo.LogError(ErrorCollectorIsBusy)
		return true, false, nil
	}
	return false, false, nil
}
