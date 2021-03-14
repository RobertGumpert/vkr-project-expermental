package tasksService

import (
	"errors"
	"fmt"
	"issue-indexer/app/models/updateTaskModel"
	"issue-indexer/pckg/requests"
	"issue-indexer/pckg/runtimeinfo"
	"net/http"
)

func (service *TasksService) sendTaskCompareIssuesInPairs(taskState *Task) error {
	if !taskState.GetExecutionStatus() {
		err := errors.New("DO NOT SEND THE RESULT TO GATE. TASK NOT COMPETED. ")
		runtimeinfo.LogError(err)
		return err
	}
	response, err := requests.POST(
		service.client,
		fmt.Sprintf(
			"%s/%s",
			service.config.GithubGateAddress,
			service.config.GithubGateEndpoints.SendResultTaskCompareIssuesInPairs,
		),
		nil,
		taskState.sendResultToGate.(updateTaskModel.UpdateTaskCompareIssuesInPairs),
	)
	if err != nil {
		runtimeinfo.LogError("SEND TO GATE WITH ERROR: ", err)
	}
	if response.StatusCode != http.StatusOK {
		runtimeinfo.LogError("SEND TO GATE WITH STATUS: ", response.StatusCode)
	}
	return err
}
