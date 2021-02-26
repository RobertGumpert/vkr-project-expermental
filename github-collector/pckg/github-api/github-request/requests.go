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

func (c *GithubClient) request(request Request, api LevelAPI) (response *http.Response, repeat bool, reset int64, err error) {
	repeat = false
	reset = int64(0)
	if request.URL == "" {
		return nil, repeat, reset, errors.New("URL is empty. ")
	}
	response, err = requests.GET(c.client, request.URL, c.addAuthHeader(request.Header))
	if err != nil {
		return nil, repeat, reset, err
	}
	time.Sleep(2 * time.Second)
	runtimeinfo.LogInfo("Request on {", request.URL, "} with status code {", response.StatusCode, "}")
	if response.StatusCode != 200 {
		if response.StatusCode == 422 || response.StatusCode == 403 {
			rate, err := c.getRateLimit()
			if err != nil {
				return nil, repeat, reset, err
			}
			switch api {
			case CORE:
				reset = rate.Resources.Core.Reset
			case SEARCH:
				reset = rate.Resources.Search.Reset
			}
			repeat = true
			return nil, repeat, reset, nil
		} else {
			return nil, repeat, reset, errors.New("Status code: " + request.URL + " = " + strconv.Itoa(response.StatusCode))
		}
	}
	return response, repeat, reset, nil
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
			responseChannel <- &Response{
				Response: nil,
				Err:      err,
			}
			c.channelTasks <- true
			c.executeTask = 0
			return
		}
		response, limitReached, resetTimeStamp, err = c.request(request, api)
		if err != nil {
			runtimeinfo.LogError("url: {", request.URL, "} err: {", err, "} ")
			responseChannel <- &Response{
				Response: nil,
				Err:      err,
			}
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
	responseChannel <- &Response{
		Response: response,
		Err:      nil,
	}
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
	var write = func(res *Response) {
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
	for {
		if numberSpentAttempts == limitNumberAttempts {
			for item := range buffer.IterBuffered() {
				write(&Response{
					URL:      item.Val.(Request).URL,
					Response: nil,
					Err:      errors.New("Number of attempts limit reached. "),
				})
			}
			break
		}
		if buffer.Count() != 0 {
			for item := range buffer.IterBuffered() {
				request := item.Val.(Request)
				response, limitReached, resetTimeStamp, err := c.request(request, api)
				if err != nil {
					runtimeinfo.LogError("url: {", request.URL, "} err: {", err, "} ")
					write(&Response{
						URL:      request.URL,
						Response: nil,
						Err:      err,
					})
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
					write(&Response{
						URL:      request.URL,
						Response: response,
						Err:      nil,
					})
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
