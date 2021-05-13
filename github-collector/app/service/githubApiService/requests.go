package githubApiService

import (
	"errors"
	"github-collector/pckg/requests"
	"github-collector/pckg/runtimeinfo"
	cmap "github.com/streamrail/concurrent-map"
	"net/http"
	"strconv"
	"time"
)

func (c *GithubClient) request(request Request, api GitHubLevelAPI) (response *http.Response, repeat bool, reset int64, err error) {
	if request.URL == "" {
		return nil, false, int64(0), errors.New("URL is empty. ")
	}
	response, err = requests.GET(c.client, request.URL, c.addAuthHeader(request.Header))
	if err != nil {
		return nil, false, int64(0), err
	}

	//  with status code {", response.StatusCode, "}
	runtimeinfo.LogInfo("Request on {", request.URL, "}")
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

func (c *GithubClient) taskOneRequest(request Request, api GitHubLevelAPI, channelNotificationRateLimit chan bool, channelGettingTaskState chan *TaskState) {
	if c.countNowExecuteTask == 0 {
		c.countNowExecuteTask = 1
	}
	runtimeinfo.LogInfo("TASK START [", request.TaskKey, "]............................................................................")
	var (
		response              *http.Response
		limitReached          bool
		err                   error
		numberSpentAttempts   int
		resetTimeStamp        int64
		writeToSignalChannel  = false
		writeToGettingChannel = func(err error) {
			runtimeinfo.LogError("url: {", request.URL, "} err: {", err, "} ")
			channelGettingTaskState <- &TaskState{
				TaskKey:       request.TaskKey,
				TaskCompleted: false,
				Responses:     []*Response{newResponse(request.TaskKey, request.URL, nil, err)},
			}
			c.tasksCompetedMessageChannel <- true
			c.countNowExecuteTask = 0
			close(channelNotificationRateLimit)
			close(channelGettingTaskState)
		}
	)
	////
	//count := 0
	////
	for {
		if numberSpentAttempts == limitNumberAttempts {
			err := errors.New("Number of attempts limit reached. ")
			writeToGettingChannel(err)
			return
		}
		response, limitReached, resetTimeStamp, err = c.request(request, api)
		if err != nil {
			writeToGettingChannel(err)
			return
		}
		////
		//count++
		//if count == 1 {
		//	limitReached = true
		//	resetTimeStamp = time.Now().Add(5 * time.Second).Unix()
		//}
		////
		if limitReached {
			if !writeToSignalChannel {
				channelNotificationRateLimit <- true
			}
			writeToSignalChannel = true
			runtimeinfo.LogInfo("Limit is reached on request [", request.URL, "] for task [", request.TaskKey, "]")
			c.freezeClient(resetTimeStamp)
			numberSpentAttempts++
			continue
		} else {
			break
		}
	}
	channelGettingTaskState <- &TaskState{
		TaskKey:       request.TaskKey,
		TaskCompleted: true,
		Responses:     []*Response{newResponse(request.TaskKey, request.URL, response, nil)},
	}
	c.tasksCompetedMessageChannel <- true
	c.countNowExecuteTask = 0
	close(channelNotificationRateLimit)
	close(channelGettingTaskState)
	runtimeinfo.LogInfo("TASK FINISH [", request.TaskKey, "]............................................................................")
}

func (c *GithubClient) taskGroupRequests(requests []Request, api GitHubLevelAPI, channelResponsesBeforeRateLimit, channelResponsesAfterRateLimit chan *TaskState) {
	if c.countNowExecuteTask == 0 {
		c.countNowExecuteTask = 1
	}
	runtimeinfo.LogInfo("TASK START [", requests[0].TaskKey, "]............................................................................")
	var (
		taskKey               = requests[0].TaskKey
		taskState             = new(TaskState)
		writeResponsesToDefer = false
		buffer                = cmap.New()
		writeResponse         = func(response *Response) {
			if taskState.TaskKey == "" {
				taskState.TaskKey = taskKey
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
					taskState.TaskCompleted = false
					if taskState.Responses != nil || len(taskState.Responses) != 0 {
						channelResponsesAfterRateLimit <- taskState
					}
					taskState = new(TaskState)
					taskState.TaskKey = taskKey
					runtimeinfo.LogInfo("Limit is reached on request [", request.URL, "] for task [", request.TaskKey, "]")
					c.freezeClient(rateLimitResetTimestamp)
					continue
				}
				if len(taskState.Responses) > 5 {
					writeResponsesToDefer = true
					taskState.TaskCompleted = false
					channelResponsesAfterRateLimit <- taskState
					taskState = new(TaskState)
					taskState.TaskKey = taskKey
				}
			}
		} else {
			break
		}
	}
	if writeResponsesToDefer {
		taskState.TaskCompleted = true
		channelResponsesAfterRateLimit <- taskState
	} else {
		taskState.TaskCompleted = true
		channelResponsesBeforeRateLimit <- taskState
	}
	c.tasksCompetedMessageChannel <- true
	c.countNowExecuteTask = 0
	close(channelResponsesAfterRateLimit)
	close(channelResponsesBeforeRateLimit)
	runtimeinfo.LogInfo("TASK FINISH [", requests[0].TaskKey, "]............................................................................")
}

func newResponse(taskKey, url string, response *http.Response, err error) *Response {
	return &Response{
		TaskKey:  taskKey,
		URL:      url,
		Response: response,
		Err:      err,
	}
}
