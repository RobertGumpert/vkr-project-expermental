package github_request

import (
	"errors"
	"net/http"
	"strings"
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

type NoWait bool
type Event func()
type TaskOneRequest func(request Request, api LevelAPI, signalChannel chan bool, responseChannel chan *Response) (Event, NoWait)
type TaskGroupRequests func(requests []Request, api LevelAPI, responsesChannel, deferResponsesChannel chan map[string]*Response) (Event, NoWait)

type GithubClient struct {
	client              *http.Client
	token               string
	isAuth              bool
	WaitRateLimitsReset bool
	maxCountTasks       int
	//
	executeTask  int
	channelTasks chan bool
	//
	tasksToOneRequest    []Event
	tasksToGroupRequests []Event
}

func NewGithubClient(token string, maxCountTasks int) (*GithubClient, error) {
	c := new(GithubClient)
	c.client = new(http.Client)
	c.WaitRateLimitsReset = false
	c.maxCountTasks = maxCountTasks
	//
	c.executeTask = 0
	c.channelTasks = make(chan bool, maxCountTasks)
	//
	c.tasksToGroupRequests = make([]Event, 0)
	c.tasksToOneRequest = make([]Event, 0)
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
	go c.nextTask()
	return c, nil
}

func (c *GithubClient) nextTask() {
	for range c.channelTasks {
		if len(c.tasksToOneRequest) != 0 {
			task := c.tasksToOneRequest[0]
			task()
			c.tasksToOneRequest = append(c.tasksToOneRequest[:0], c.tasksToOneRequest[0+1:]...)
			continue
		}
		if len(c.tasksToGroupRequests) != 0 {
			task := c.tasksToGroupRequests[0]
			task()
			c.tasksToGroupRequests = append(c.tasksToGroupRequests[:0], c.tasksToGroupRequests[0+1:]...)
			continue
		}
	}
}

func (c *GithubClient) AddOneRequest() (TaskOneRequest, error) {
	if len(c.tasksToOneRequest) == c.maxCountTasks {
		return nil, errors.New("Limit on the number of tasks has been reached. ")
	}
	all := len(c.tasksToOneRequest) + len(c.tasksToGroupRequests)
	if all == c.maxCountTasks {
		return nil, errors.New("Limit on the number of tasks has been reached. ")
	}
	return func(request Request, api LevelAPI, signalChannel chan bool, responseChannel chan *Response) (Event, NoWait) {
		var task = func() {
			c.taskOneRequest(request, api, signalChannel, responseChannel)
		}
		if c.executeTask == 0 {
			return task, true
		}
		if len(c.tasksToOneRequest) != 0 || c.executeTask == 1 {
			c.tasksToOneRequest = append(c.tasksToOneRequest, task)
		}
		return task, false
	}, nil
}

func (c *GithubClient) AddGroupRequests() (TaskGroupRequests, error) {
	if len(c.tasksToGroupRequests) == c.maxCountTasks {
		return nil, errors.New("Limit on the number of tasks has been reached. ")
	}
	all := len(c.tasksToOneRequest) + len(c.tasksToGroupRequests)
	if all == c.maxCountTasks {
		return nil, errors.New("Limit on the number of tasks has been reached. ")
	}
	return func(requests []Request, api LevelAPI, responsesChannel, deferResponsesChannel chan map[string]*Response) (Event, NoWait) {
		var task = func() {
			c.taskGroupRequests(requests, api, responsesChannel, deferResponsesChannel)
		}
		if c.executeTask == 0 {
			return task, true
		}
		if len(c.tasksToGroupRequests) != 0 || c.executeTask == 1 {
			c.tasksToGroupRequests = append(c.tasksToGroupRequests, task)
		}
		return task, false
	}, nil
}
