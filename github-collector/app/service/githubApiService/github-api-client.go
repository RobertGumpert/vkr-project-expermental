package githubApiService

import (
	"errors"
	"github-collector/pckg/runtimeinfo"
	"net/http"
	"strings"
	"sync"
)

type TaskIndexInQueue int

type QueueIsFilled bool

// Сигнализирует о том, что необходимо запустить задачу (Task),
// не дожидаясь итератора задач (c *GithubClient) scanTask().
type IsRunTaskNow bool

// Функция, которая запускает задачу (Task),
// в теле которой, выполняется запуск уже настроенной задачи.
type TaskLaunchFunction func()

// Функция, которая настраивает задачу
// выполнения одного запроса к GitHub.
// Возвращает функцию запуска задачи - TaskLaunchFunction,
// которую необходимо запускать если значение IsRunTaskNow = true.
//
// Аргументами FunctionInsertOneRequest являются параметры запроса:
// 	* request Request 		  		 - содержит URL и HEADER запроса.
// 	* api GitHubLevelAPI 					 - уровень API GitHub (Core, Search).
// 	* signalChannel chan bool 		 - канал для передачи сообщения,
// 									   о том что Rate Limit достигнут
// 									   и не следует ждать завершения задачи,
// 									   так как она завершится позже и результат будет
// 									   записан в responseChannel chan *Response.
// 	* responseChannel chan *Response - канал передачи ответа от API GitHub.
type FunctionInsertOneRequest func(
	request Request,
	api GitHubLevelAPI,
	channelNotificationRateLimit chan bool,
	channelGettingTaskState chan *TaskState,
) (TaskLaunchFunction, IsRunTaskNow, TaskIndexInQueue)

// Функция, которая настраивает задачу
// выполнения группы запросов к GitHub.
// Возвращает функцию запуска задачи - TaskLaunchFunction,
// которую необходимо запускать если значение IsRunTaskNow = true.
//
// Аргументами FunctionInsertGroupRequests являются параметры запроса:
// 	* request Request 		  		                  - содержит URL и HEADER запроса.
// 	* api GitHubLevelAPI 					                  - уровень API GitHub (Core, Search).
// 	* responsesChannel chan map[string]*Response 	  - канал передачи ответов от API GitHub,
// 														передаются ответы на уже выполненые запросы,
// 														без достижения Rate Limit.
// 	* deferResponsesChannel chan map[string]*Response - канал передачи ответов от API GitHub,
//														передаются ответы на уже выполненые запросы,
//														до момента достижения Rate Limit,
//														а отсальные ответы будут переданы позже,
// 									   					соответсвенно не следует ждать завершения задачи.
type FunctionInsertGroupRequests func(
	requests []Request,
	api GitHubLevelAPI,
	channelResponseBeforeRateLimit,
	channelResponseAfterRateLimit chan *TaskState,
) (TaskLaunchFunction, IsRunTaskNow, TaskIndexInQueue)



type GithubClient struct {
	mx *sync.Mutex
	//
	client              *http.Client
	token               string
	isAuth              bool
	WaitRateLimitsReset bool
	maxCountTasks       int
	//
	countNowExecuteTask         int
	tasksCompetedMessageChannel chan bool
	//
	tasksToOneRequest    []TaskLaunchFunction
	tasksToGroupRequests []TaskLaunchFunction
}

func NewGithubClient(token string, maxCountTasks int) (*GithubClient, error) {
	c := new(GithubClient)
	c.client = new(http.Client)
	c.WaitRateLimitsReset = false
	c.maxCountTasks = maxCountTasks
	c.mx = new(sync.Mutex)
	c.countNowExecuteTask = 0
	c.tasksCompetedMessageChannel = make(chan bool, maxCountTasks)
	c.tasksToGroupRequests = make([]TaskLaunchFunction, 0)
	c.tasksToOneRequest = make([]TaskLaunchFunction, 0)
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
	go c.scanTask()
	return c, nil
}

func (c *GithubClient) scanTask() {
	for range c.tasksCompetedMessageChannel {
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

func (c *GithubClient) GetState() (error, int) {
	all := len(c.tasksToOneRequest) + len(c.tasksToGroupRequests)
	if all == c.maxCountTasks {
		return errors.New("Limit on the number of tasks has been reached. "), all
	}
	return nil, all
}

// Для создания новой задачи на выполнение одного запроса
// к GitHub, необходимо сначала проверить состояние очереди:
// 	* если она переполнена возвращается ошибка.
// 	* если место в очереди есть, возвращай функцию
// 	  настройки запроса к GitHub (FunctionInsertOneRequest).
//
// Аргументами FunctionInsertOneRequest являются параметры запроса:
// 	* request Request 		  		 - содержит URL и HEADER запроса.
// 	* api GitHubLevelAPI 					 - уровень API GitHub (Core, Search).
// 	* signalChannel chan bool 		 - канал для передачи сообщения,
// 									   о том что Rate Limit достигнут
// 									   и не следует ждать завершения задачи,
// 									   так как она завершится позже и результат будет
// 									   записан в responseChannel chan *Response.
// 	* responseChannel chan *Response - канал передачи ответа от API GitHub.
//
//
func (c *GithubClient) MakeFunctionInsertOneRequest(makeTaskDefer bool) (FunctionInsertOneRequest, QueueIsFilled) {
	if !makeTaskDefer {
		if len(c.tasksToOneRequest) == c.maxCountTasks {
			return nil, true
		}
		all := len(c.tasksToOneRequest) + len(c.tasksToGroupRequests)
		if all == c.maxCountTasks {
			return nil, true
		}
	}
	return func(request Request, api GitHubLevelAPI, signalChannel chan bool, taskStateChannel chan *TaskState) (TaskLaunchFunction, IsRunTaskNow, TaskIndexInQueue) {
		var runTask = func() {
			c.taskOneRequest(request, api, signalChannel, taskStateChannel)
		}
		return c.addTask(runTask, false)
	}, false
}

// Для создания новой задачи на выполнение группы запросов
// к GitHub, необходимо сначала проверить состояние очереди:
// 	* если она переполнена возвращается ошибка.
// 	* если место в очереди есть, возвращай функцию
// 	  настройки запроса к GitHub (FunctionInsertGroupRequests).
//
// Аргументами FunctionInsertGroupRequests являются параметры запроса:
// 	* request Request 		  		                  - содержит URL и HEADER запроса.
// 	* api GitHubLevelAPI 					                  - уровень API GitHub (Core, Search).
// 	* responsesChannel chan map[string]*Response 	  - канал передачи ответов от API GitHub,
// 														передаются ответы на уже выполненые запросы,
// 														без достижения Rate Limit.
// 	* deferResponsesChannel chan map[string]*Response - канал передачи ответов от API GitHub,
//														передаются ответы на уже выполненые запросы,
//														до момента достижения Rate Limit,
//														а отсальные ответы будут переданы позже,
// 									   					соответсвенно не следует ждать завершения задачи.
func (c *GithubClient) MakeFunctionInsertGroupRequests(makeTaskDefer bool) (FunctionInsertGroupRequests, QueueIsFilled) {
	if !makeTaskDefer {
		if len(c.tasksToGroupRequests) == c.maxCountTasks {
			return nil, true
		}
		all := len(c.tasksToOneRequest) + len(c.tasksToGroupRequests)
		if all == c.maxCountTasks {
			return nil, true
		}
	}
	return func(requests []Request, api GitHubLevelAPI, taskStateChannel, deferTaskStateChannel chan *TaskState) (TaskLaunchFunction, IsRunTaskNow, TaskIndexInQueue) {
		var runTask = func() {
			c.taskGroupRequests(requests, api, taskStateChannel, deferTaskStateChannel)
		}
		return c.addTask(runTask, true)
	}, false
}

func (c *GithubClient) addTask(runTask func(), isGroup bool) (TaskLaunchFunction, IsRunTaskNow, TaskIndexInQueue) {
	c.mx.Lock()
	defer c.mx.Unlock()
	//
	if c.countNowExecuteTask == 0 {
		c.countNowExecuteTask++
		runtimeinfo.LogInfo("Add new request/s as runnable.")
		return runTask, true, 0
	}
	if len(c.tasksToGroupRequests) != 0 || c.countNowExecuteTask == 1 {
		runtimeinfo.LogInfo("Add new request/s as defer.")
		if isGroup {
			c.tasksToGroupRequests = append(c.tasksToGroupRequests, runTask)
		} else {
			c.tasksToOneRequest = append(c.tasksToOneRequest, runTask)
		}
	}
	return runTask, false, TaskIndexInQueue(len(c.tasksToGroupRequests) - 1)
}