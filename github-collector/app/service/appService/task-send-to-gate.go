package appService

import (
	"encoding/json"
	"errors"
	"fmt"
	"github-collector/app/service/githubApiService"
	"github-collector/pckg/requests"
	"github-collector/pckg/runtimeinfo"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

func (service *AppService) repeatResponsesToGithubGate() {
	for {
		runtime.Gosched()
		runtimeinfo.LogInfo("REPEATED TASK RESPONSES...")
		for index, repeatedResponse := range service.repeatedResponses {
			if err := repeatedResponse(); err == nil {
				service.repeatedResponses = append(service.repeatedResponses[:index], service.repeatedResponses[index+1:]...)
			}
		}
		runtimeinfo.LogInfo("REPEATED TASK RESPONSES SLEEP...")
		time.Sleep(10 * time.Minute)
	}
}

func (service *AppService) doResponseToGate(body interface{}, endpoint string) (err error) {
	url := fmt.Sprintf("%s%s", service.config.GithubGateAddress, endpoint)
	response, err := requests.POST(
		service.client,
		url,
		nil,
		body,
	)
	if err != nil {
		runtimeinfo.LogError("[", url, "]", err)
		return err
	}
	if response.StatusCode != http.StatusOK {
		err := errors.New("[" + url + "] status: " + strconv.Itoa(response.StatusCode))
		runtimeinfo.LogError(err)
		return err
	}
	return nil
}

func (service *AppService) sendTaskRepositoriesDescriptions(taskState *githubApiService.TaskState) {
	var (
		sendBody        = JsonUpdateTaskRepositoriesDescriptions{}
		executionStatus = JsonExecutionStatus{
			TaskKey:       taskState.TaskKey,
			TaskCompleted: taskState.TaskCompleted,
		}
		dataModels = make([]RepositoryDataModel, 0)
		taskKey    string
		doResponse = func() (err error) {
			runtimeinfo.LogInfo("DO SEND UPDATES FOR TASK: [", sendBody.ExecutionTaskStatus.TaskKey, "]")
			return service.doResponseToGate(
				&sendBody,
				service.config.GithubGateEndpoints.SendResponseTaskRepositoriesDescriptions,
			)
		}
	)
	if taskState.Responses != nil {
		for _, response := range taskState.Responses {
			if taskKey == "" {
				taskKey = response.TaskKey
			}
			dataModel := RepositoryDataModel{}
			err := json.NewDecoder(response.Response.Body).Decode(&dataModel)
			if err != nil {
				runtimeinfo.LogError(err)
			}
			dataModels = append(dataModels, dataModel)
		}
	}
	sendBody.Repositories = dataModels
	sendBody.ExecutionTaskStatus = executionStatus
	err := doResponse()
	if err != nil {
		service.repeatedResponses = append(service.repeatedResponses, doResponse)
	}
	return
}

func (service *AppService) sendTaskRepositoryIssues(taskState *githubApiService.TaskState) {
	var (
		sendBody        = JsonUpdateTaskRepositoryIssues{}
		executionStatus = JsonExecutionStatus{
			TaskKey:       taskState.TaskKey,
			TaskCompleted: taskState.TaskCompleted,
		}
		dataModels = make([]IssueDataModel, 0)
		taskKey    string
		doResponse = func() (err error) {
			runtimeinfo.LogInfo("DO SEND UPDATES FOR TASK: [", sendBody.ExecutionTaskStatus.TaskKey, "]")
			return service.doResponseToGate(
				&sendBody,
				service.config.GithubGateEndpoints.SendResponseTaskRepositoryIssues,
			)
		}
	)
	if taskState.Responses != nil {
		for index, response := range taskState.Responses {
			if response == nil {
				runtimeinfo.LogError("response equals nil :[", index, "];")
				continue
			}
			if response.Response == nil {
				runtimeinfo.LogError("response body equals nil :[", index, "];")
				continue
			}
			if taskKey == "" {
				taskKey = response.TaskKey
			}
			body, err := ioutil.ReadAll(response.Response.Body)
			if err != nil {
				runtimeinfo.LogError(err)
				continue
			}
			var list []IssueDataModel
			err = json.Unmarshal(body, &list)
			if err != nil {
				runtimeinfo.LogError(err)
				continue
			}
			dataModels = append(dataModels, list...)
		}
	}
	sendBody.Issues = dataModels
	sendBody.ExecutionTaskStatus = executionStatus
	err := doResponse()
	if err != nil {
		service.repeatedResponses = append(service.repeatedResponses, doResponse)
	}
	return
}

func (service *AppService) sendTaskRepositoriesByKeyWord(taskState *githubApiService.TaskState) {
	var (
		sendBody        = JsonUpdateTaskRepositoriesByKeyWord{}
		executionStatus = JsonExecutionStatus{
			TaskKey:       taskState.TaskKey,
			TaskCompleted: taskState.TaskCompleted,
		}
		dataModels = make([]RepositoryDataModel, 0)
		taskKey    string
		doResponse = func() (err error) {
			runtimeinfo.LogInfo("DO SEND UPDATES FOR TASK: [", sendBody.ExecutionTaskStatus.TaskKey, "]")
			return service.doResponseToGate(
				&sendBody,
				service.config.GithubGateEndpoints.SendResponseTaskRepositoriesByKeyWord,
			)
		}
	)
	if taskState.Responses != nil {
		response := taskState.Responses[0]
		if taskKey == "" {
			taskKey = response.TaskKey
		}
		dataModel := RepositoriesByKeyWordDataModel{}
		err := json.NewDecoder(response.Response.Body).Decode(&dataModel)
		if err != nil {
			runtimeinfo.LogError(err)
		}
		dataModels = append(dataModels, dataModel.Items...)
	}
	sendBody.Repositories = dataModels
	sendBody.ExecutionTaskStatus = executionStatus
	err := doResponse()
	if err != nil {
		service.repeatedResponses = append(service.repeatedResponses, doResponse)
	}
	return
}
