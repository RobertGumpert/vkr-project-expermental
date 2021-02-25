package github_request

import (
	"errors"
	"github-collector/pckg/requests"
	"github-collector/pckg/runtimeinfo"
	cmap "github.com/streamrail/concurrent-map"
	"net/http"
	"strings"
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

type rateLimitInfo struct {
	Resources struct {
		Core struct {
			Used  int64 `json:"used"`
			Reset int64 `json:"reset"`
		} `json:"core"`
		Search struct {
			Used  int64 `json:"used"`
			Reset int64 `json:"reset"`
		} `json:"search"`
	} `json:"resources"`
	Rate struct {
		Used  int64 `json:"used"`
		Reset int64 `json:"reset"`
	} `json:"rate"`
}

type GithubClient struct {
	client              *http.Client
	token               string
	isAuth              bool
	WaitRateLimitsReset bool
	waitChannel         chan bool
}

type Response struct {
	Response *http.Response
	Err      error
}

type Request struct {
	URL                 string
	Header              map[string]string
	numberSpentAttempts int
}

func NewGithubClient(token string) (*GithubClient, error) {
	c := new(GithubClient)
	c.client = new(http.Client)
	c.waitChannel = make(chan bool)
	c.WaitRateLimitsReset = false
	//
	if token != "" {
		token = strings.Join([]string{
			"token",
			token,
		}, " ")
		c.token = token
		err := c.auth()
		if err != nil {
			return nil, err
		}
		c.isAuth = true
	} else {
		c.isAuth = false
	}
	return c, nil
}

func (c *GithubClient) auth() error {
	response, err := requests.GET(c.client, authURL, map[string]string{
		"Authorization": c.token,
	})
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("Status code not 200. ")
	}
	return nil
}

func (c *GithubClient) addAuthHeader(header map[string]string) map[string]string {
	if header == nil && c.isAuth {
		header = map[string]string{
			"Authorization": c.token,
		}
	}
	if header != nil && c.isAuth {
		header["Authorization"] = c.token
	}
	return header
}

func (c *GithubClient) getRateLimit() (*rateLimitInfo, error) {
	response, err := requests.GET(c.client, rateLimitURL, c.addAuthHeader(nil))
	if err != nil {
		return nil, err
	}
	var rate *rateLimitInfo
	if err := requests.Deserialize(&rate, response); err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New("Status code not 200. ")
	}
	return rate, nil
}

func (c *GithubClient) sleep(reset int64) {
	c.WaitRateLimitsReset = true
	timeNow := time.Now()
	timeReset := time.Unix(reset, int64(0))
	when := timeNow.Sub(timeReset)
	runtimeinfo.LogInfo("CLIENT FREEZE ON ", when, "...")
	time.Sleep(when)
	c.waitChannel <- true
	c.WaitRateLimitsReset = false
	runtimeinfo.LogInfo("CLIENT UNFREEZE.")
}


func (c *GithubClient) writeToSignalChanel(signalChannel chan interface{}, val interface{}) {
	signalChannel <- val
}

func (c *GithubClient) do(request Request, api LevelAPI) (response *http.Response, repeat bool, reset int64, err error) {
	repeat = false
	reset = int64(0)
	if request.URL == "" {
		return nil, repeat, reset, errors.New("URL is empty. ")
	}
	response, err = requests.GET(c.client, request.URL, c.addAuthHeader(request.Header))
	if err != nil {
		return nil, repeat, reset, err
	}
	runtimeinfo.LogInfo("Request on {", request.URL, "} with status code {", response.StatusCode, "}")
	if response.StatusCode != 200 {
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
	}
	return nil, repeat, reset, nil
}


func (c *GithubClient) Request(request Request, api LevelAPI, signalChannel chan bool, responses chan *Response) {
	var (
		response            *http.Response
		doRepeatRequest     bool
		err                 error
		numberSpentAttempts int
		reset int64
	)
	for {
		if numberSpentAttempts == limitNumberAttempts {
			err := errors.New("Ended numberSpentAttempts. ")
			runtimeinfo.LogError("url: {", request.URL, "} err: {", err, "} ")
			responses <- &Response{
				Response: nil,
				Err:      err,
			}
			return
		}
		if c.WaitRateLimitsReset {
			signalChannel <- true
			runtimeinfo.LogInfo("Wait when client unfreeze. URL: {", request.URL, "}")
			<-c.waitChannel
			runtimeinfo.LogInfo("client unfreeze. URL: {", request.URL, "}")
		}
		response, doRepeatRequest, reset, err = c.do(request, api)
		if err != nil {
			runtimeinfo.LogError("url: {", request.URL, "} err: {", err, "} ")
			responses <- &Response{
				Response: nil,
				Err:      err,
			}
			return
		}
		if doRepeatRequest {
			signalChannel <- true
			c.sleep(reset)
			runtimeinfo.LogInfo("Repeat request on {", request.URL, "} ")
			numberSpentAttempts++
			continue
		} else {
			break
		}
	}
	responses <- &Response{
		Response: response,
		Err:      nil,
	}
}

func (c *GithubClient) Requests(requests []Request, api LevelAPI, mainChannelResponses, lazyChannelResponses chan map[string]*Response) {
	var (
		mainMapResponses = make(map[string]*Response)
		lazyMapResponses = make(map[string]*Response)
		lazyMapWriting   = false
		buffer           = cmap.New()
	)
	if c.WaitRateLimitsReset {
		mainChannelResponses <- mainMapResponses
		runtimeinfo.LogInfo("Wait when client unfreeze. URLs.")
		<-c.waitChannel
		runtimeinfo.LogInfo("client unfreeze. URLs.")
	}
	for _, request := range requests {
		buffer.Set(request.URL, request)
	}
	for {
		if buffer.Count() != 0 {
			for item := range buffer.IterBuffered() {
				request := item.Val.(Request)
				if request.numberSpentAttempts == limitNumberAttempts {
					err := errors.New("Ended numberSpentAttempts. ")
					runtimeinfo.LogError("url: {", request.URL, "} err: {", err, "} ")
					lazyMapResponses[request.URL] = &Response{
						Response: nil,
						Err:      err,
					}
					buffer.Remove(request.URL)
					return
				}
				response, doRepeatRequest, reset, err := c.do(request, api)
				if err != nil {
					runtimeinfo.LogError("url: {", request.URL, "} err: {", err, "} ")
					mainMapResponses[request.URL] = &Response{
						Response: nil,
						Err:      err,
					}
					buffer.Remove(request.URL)
					continue
				}
				if doRepeatRequest {
					lazyMapWriting = true
					mainChannelResponses <- mainMapResponses
					c.sleep(reset)
					request.numberSpentAttempts++
					runtimeinfo.LogInfo("Repeat request on {", request.URL, "} ")
					continue
				}
				if response != nil {
					if lazyMapWriting {
						lazyMapResponses[request.URL] = &Response{
							Response: response,
							Err:      nil,
						}
					} else {
						mainMapResponses[request.URL] = &Response{
							Response: response,
							Err:      nil,
						}
					}
					buffer.Remove(request.URL)
				}
			}
		} else {
			break
		}
	}
	if lazyMapWriting {
		lazyChannelResponses <- lazyMapResponses
	} else {
		mainChannelResponses <- mainMapResponses
	}
}
