package github_request

import (
	"errors"
	"github-collector/pckg/requests"
	"github-collector/pckg/runtimeinfo"
	cmap "github.com/streamrail/concurrent-map"
	"net/http"
	"strconv"
	"time"
)

type GitHubLevelAPI uint64

const (
	CORE                GitHubLevelAPI = 0
	SEARCH              GitHubLevelAPI = 1
	maxCoreRequests     uint64         = 5000
	maxSearchRequests   uint64         = 30
	limitNumberAttempts int            = 5
	authURL                            = "https://api.github.com/user"
	rateLimitURL                       = "https://api.github.com/rate_limit"
)

//
//----------------------------------------------------------------------------------------------------------------------
//

type TaskState struct {
	TaskKey         string
	ExecutionStatus bool
	Responses       []*Response
}

type Response struct {
	TaskKey  string
	URL      string
	Response *http.Response
	Err      error
}

type Request struct {
	TaskKey             string
	URL                 string
	Header              map[string]string
	numberSpentAttempts int
}

//
//----------------------------------------------------------------------------------------------------------------------
//

func (c *GithubClient) request(request Request, api GitHubLevelAPI) (response *http.Response, repeat bool, reset int64, err error) {
	if request.URL == "" {
		return nil, false, int64(0), errors.New("URL is empty. ")
	}
	response, err = requests.GET(c.client, request.URL, c.addAuthHeader(request.Header))
	if err != nil {
		return nil, false, int64(0), err
	}
	runtimeinfo.LogInfo("Request on {", request.URL, "} with status code {", response.StatusCode, "}")
	if response.StatusCode != 200 {
		if response.StatusCode == 422 || response.StatusCode == 403 {
			rate, err := c.getRateLimit()
			if err != nil {
				return nil, false, int64(0), err
			}
			switch api {
			case CORE:
				reset = rate.Resources.Core.Reset
			case SEARCH:
				reset = rate.Resources.Search.Reset
			}
			return nil, true, reset, nil
		} else {
			return nil, false, int64(0), errors.New("Status code: " + request.URL + " = " + strconv.Itoa(response.StatusCode))
		}
	}
	time.Sleep(2 * time.Second)
	return response, false, int64(0), nil
}

func (c *GithubClient) taskOneRequest(request Request, api GitHubLevelAPI, signalChannel chan bool, taskStateChannel chan *TaskState) {
	c.countNowExecuteTask = 1
	runtimeinfo.LogInfo("TASK START [", request.TaskKey, "]............................................................................")
	var (
		response             *http.Response
		limitReached         bool
		err                  error
		numberSpentAttempts  int
		resetTimeStamp       int64
		writeToSignalChannel = false
		completeTask         = func(err error) {
			runtimeinfo.LogError("url: {", request.URL, "} err: {", err, "} ")
			taskStateChannel <- &TaskState{
				TaskKey:         request.TaskKey,
				ExecutionStatus: false,
				Responses:       []*Response{newResponse(request.TaskKey, request.URL, nil, err)},
			}
			c.tasksCompetedMessageChannel <- true
			c.countNowExecuteTask = 0
		}
	)
	for {
		if numberSpentAttempts == limitNumberAttempts {
			err := errors.New("Number of attempts limit reached. ")
			completeTask(err)
			return
		}
		response, limitReached, resetTimeStamp, err = c.request(request, api)
		if err != nil {
			completeTask(err)
			return
		}
		if limitReached {
			if !writeToSignalChannel {
				signalChannel <- true
			}
			writeToSignalChannel = true
			c.freezeClient(resetTimeStamp)
			runtimeinfo.LogInfo("Repeat request on {", request.URL, "} ")
			numberSpentAttempts++
			continue
		} else {
			break
		}
	}
	taskStateChannel <- &TaskState{
		TaskKey:         request.TaskKey,
		ExecutionStatus: true,
		Responses:       []*Response{newResponse(request.TaskKey, request.URL, response, nil)},
	}
	c.tasksCompetedMessageChannel <- true
	c.countNowExecuteTask = 0
	close(signalChannel)
	close(taskStateChannel)
	runtimeinfo.LogInfo("TASK START [", request.TaskKey, "]............................................................................")
}

func (c *GithubClient) taskGroupRequests(requests []Request, api GitHubLevelAPI, taskStateChannel, deferTaskStateChannel chan *TaskState) {
	c.countNowExecuteTask = 1
	runtimeinfo.LogInfo("TASK START [",requests[0].TaskKey,"]............................................................................")
	var (
		taskState             = new(TaskState)
		writeResponsesToDefer = false
		buffer                = cmap.New()
		writeResponse         = func(response *Response) {
			if taskState.TaskKey == "" {
				taskState.TaskKey = requests[0].TaskKey
			}
			if taskState.Responses == nil {
				taskState.Responses = make([]*Response, 0)
			}
			if response.Response == nil {
				response.Err = errors.New("Response is nil. ")
			}
			taskState.Responses = append(taskState.Responses, response)
			return
		}
	)
	for _, request := range requests {
		buffer.Set(request.URL, request)
	}
	for {
		if buffer.Count() != 0 {
			for item := range buffer.IterBuffered() {
				request := item.Val.(Request)
				httpResponse, limitReached, rateLimitResetTimestamp, err := c.request(request, api)
				if limitReached && !writeResponsesToDefer {
					writeResponsesToDefer = true
				}
				writeResponse(
					newResponse(
						request.TaskKey,
						request.URL,
						httpResponse,
						err,
					),
				)
				if limitReached == false {
					buffer.Remove(request.URL)
				}
				if limitReached && writeResponsesToDefer {
					taskState.ExecutionStatus = false
					if taskState.Responses != nil || len(taskState.Responses) != 0 {
						deferTaskStateChannel <- taskState
					}
					taskState = new(TaskState)
					runtimeinfo.LogInfo("Repeat requests...")
					c.freezeClient(rateLimitResetTimestamp)
					continue
				}
				if len(taskState.Responses) > 5 {
					writeResponsesToDefer = true
					taskState.ExecutionStatus = false
					deferTaskStateChannel <- taskState
					taskState = new(TaskState)
				}
			}
		} else {
			break
		}
	}
	if writeResponsesToDefer {
		taskState.ExecutionStatus = true
		deferTaskStateChannel <- taskState
	} else {
		taskState.ExecutionStatus = true
		taskStateChannel <- taskState
	}
	c.tasksCompetedMessageChannel <- true
	c.countNowExecuteTask = 0
	close(deferTaskStateChannel)
	close(taskStateChannel)
	runtimeinfo.LogInfo("TASK FINISH [",requests[0].TaskKey,"]............................................................................")
}

func newResponse(taskKey, url string, response *http.Response, err error) *Response {
	return &Response{
		TaskKey:  taskKey,
		URL:      url,
		Response: response,
		Err:      err,
	}
}
