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

type LevelAPI uint64

const (
	CORE                LevelAPI = 0
	SEARCH              LevelAPI = 1
	maxCoreRequests     uint64   = 5000
	maxSearchRequests   uint64   = 30
	limitNumberAttempts int      = 5
	authURL                      = "https://api.github.com/user"
	rateLimitURL                 = "https://api.github.com/rate_limit"
)

//
//----------------------------------------------------------------------------------------------------------------------
//

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

func (c *GithubClient) request(request Request, api LevelAPI) (response *http.Response, repeat bool, reset int64, err error) {
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

func (c *GithubClient) taskOneRequest(request Request, api LevelAPI, signalChannel chan bool, responseChannel chan *Response) {
	c.executeTask = 1
	runtimeinfo.LogInfo("TASK START............................................................................")
	var (
		response            *http.Response
		limitReached        bool
		err                 error
		numberSpentAttempts int
		resetTimeStamp      int64
	)
	for {
		if numberSpentAttempts == limitNumberAttempts {
			err := errors.New("Number of attempts limit reached. ")
			runtimeinfo.LogError("url: {", request.URL, "} err: {", err, "} ")
			responseChannel <- newResponse(request.TaskKey, request.URL, nil, err)
			c.channelTasks <- true
			c.executeTask = 0
			return
		}
		response, limitReached, resetTimeStamp, err = c.request(request, api)
		if err != nil {
			runtimeinfo.LogError("url: {", request.URL, "} err: {", err, "} ")
			responseChannel <- newResponse(request.TaskKey, request.URL, nil, err)
			c.channelTasks <- true
			c.executeTask = 0
			return
		}
		if limitReached {
			signalChannel <- true
			c.freezeClient(resetTimeStamp)
			runtimeinfo.LogInfo("Repeat request on {", request.URL, "} ")
			numberSpentAttempts++
			continue
		} else {
			break
		}
	}
	responseChannel <- newResponse(request.TaskKey, request.URL, response, nil)
	c.channelTasks <- true
	c.executeTask = 0
	runtimeinfo.LogInfo("TASK FINISH............................................................................")
}

func (c *GithubClient) taskGroupRequests(requests []Request, api LevelAPI, responsesChannel, deferResponsesChannel chan map[string]*Response) {
	c.executeTask = 1
	runtimeinfo.LogInfo("TASK START............................................................................")
	var (
		responses           = make(map[string]*Response)
		deferResponses      = make(map[string]*Response)
		writeDefer          = false
		numberSpentAttempts = 0
		buffer              = cmap.New()
	)
	//
	var writeResponseToMap = func(res *Response) {
		if writeDefer == true {
			deferResponses[res.URL] = res
		}
		if writeDefer == false {
			responses[res.URL] = res
		}
		buffer.Remove(res.URL)
		return
	}
	//
	for _, request := range requests {
		buffer.Set(request.URL, request)
	}
	//
	for {
		if numberSpentAttempts == limitNumberAttempts {
			for item := range buffer.IterBuffered() {
				err := errors.New("Number of attempts limit reached. ")
				runtimeinfo.LogError("url: {", item.Val.(Request).URL, "} err: {", err, "} ")
				writeResponseToMap(newResponse(item.Val.(Request).TaskKey, item.Val.(Request).URL, nil, err))
			}
			break
		}
		if buffer.Count() != 0 {
			for item := range buffer.IterBuffered() {
				request := item.Val.(Request)
				response, limitReached, resetTimeStamp, err := c.request(request, api)
				if err != nil {
					runtimeinfo.LogError("url: {", request.URL, "} err: {", err, "} ")
					writeResponseToMap(newResponse(request.TaskKey, request.URL, nil, err))
					continue
				}
				if limitReached {
					numberSpentAttempts++
					writeDefer = true
					if numberSpentAttempts == 1 {
						deferResponsesChannel <- responses
					}
					runtimeinfo.LogInfo("Repeat requests...")
					c.freezeClient(resetTimeStamp)
					continue
				}
				if response != nil {
					writeResponseToMap(newResponse(request.TaskKey, request.URL, response, nil))
				}
			}
		} else {
			break
		}
	}
	if writeDefer {
		deferResponsesChannel <- deferResponses
	} else {
		responsesChannel <- responses
	}
	c.channelTasks <- true
	c.executeTask = 0
	runtimeinfo.LogInfo("TASK FINISH............................................................................")
}

func newResponse(taskKey, url string, response *http.Response, err error) *Response {
	return &Response{
		TaskKey:  taskKey,
		URL:      url,
		Response: response,
		Err:      err,
	}
}