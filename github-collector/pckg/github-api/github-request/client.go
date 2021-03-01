package github_request

import (
	"errors"
	"net/http"
	"strings"
)

// Сигнализирует о том, что необходимо запустить задачу (Task),
// не дожидаясь итератора задач (c *GithubClient) scanTask().
type NoWait bool

// Функция, которая запускает задачу (Task),
// в теле которой, выполняется запуск уже настроенной задачи.
type RunTask func()

// Функция, которая настраивает задачу
// выполнения одного запроса к GitHub.
// Возвращает функцию запуска задачи - RunTask,
// которую необходимо запускать если значение NoWait = true.
//
// Аргументами TaskOneRequest являются параметры запроса:
// 	* request Request 		  		 - содержит URL и HEADER запроса.
// 	* api LevelAPI 					 - уровень API GitHub (Core, Search).
// 	* signalChannel chan bool 		 - канал для передачи сообщения,
// 									   о том что Rate Limit достигнут
// 									   и не следует ждать завершения задачи,
// 									   так как она завершится позже и результат будет
// 									   записан в responseChannel chan *Response.
// 	* responseChannel chan *Response - канал передачи ответа от API GitHub.
type TaskOneRequest func(request Request, api LevelAPI, signalChannel chan bool, taskStateChannel chan *TaskState) (RunTask, NoWait, int)

// Функция, которая настраивает задачу
// выполнения группы запросов к GitHub.
// Возвращает функцию запуска задачи - RunTask,
// которую необходимо запускать если значение NoWait = true.
//
// Аргументами TaskGroupRequests являются параметры запроса:
// 	* request Request 		  		                  - содержит URL и HEADER запроса.
// 	* api LevelAPI 					                  - уровень API GitHub (Core, Search).
// 	* responsesChannel chan map[string]*Response 	  - канал передачи ответов от API GitHub,
// 														передаются ответы на уже выполненые запросы,
// 														без достижения Rate Limit.
// 	* deferResponsesChannel chan map[string]*Response - канал передачи ответов от API GitHub,
//														передаются ответы на уже выполненые запросы,
//														до момента достижения Rate Limit,
//														а отсальные ответы будут переданы позже,
// 									   					соответсвенно не следует ждать завершения задачи.
type TaskGroupRequests func(requests []Request, api LevelAPI, taskStateChannel, deferTaskStateChannel chan *TaskState) (RunTask, NoWait, int)

type GithubClient struct {
	client              *http.Client
	token               string
	isAuth              bool
	WaitRateLimitsReset bool
	maxCountTasks       int
	//
	countNowExecuteTask         int
	tasksCompetedMessageChannel chan bool
	//
	tasksToOneRequest    []RunTask
	tasksToGroupRequests []RunTask
}

func NewGithubClient(token string, maxCountTasks int) (*GithubClient, error) {
	c := new(GithubClient)
	c.client = new(http.Client)
	c.WaitRateLimitsReset = false
	c.maxCountTasks = maxCountTasks
	c.countNowExecuteTask = 0
	c.tasksCompetedMessageChannel = make(chan bool, maxCountTasks)
	c.tasksToGroupRequests = make([]RunTask, 0)
	c.tasksToOneRequest = make([]RunTask, 0)
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
// 	  настройки запроса к GitHub (TaskOneRequest).
//
// Аргументами TaskOneRequest являются параметры запроса:
// 	* request Request 		  		 - содержит URL и HEADER запроса.
// 	* api LevelAPI 					 - уровень API GitHub (Core, Search).
// 	* signalChannel chan bool 		 - канал для передачи сообщения,
// 									   о том что Rate Limit достигнут
// 									   и не следует ждать завершения задачи,
// 									   так как она завершится позже и результат будет
// 									   записан в responseChannel chan *Response.
// 	* responseChannel chan *Response - канал передачи ответа от API GitHub.
//
//
func (c *GithubClient) AddOneRequest(reserved bool) (TaskOneRequest, error) {
	if !reserved {
		if len(c.tasksToOneRequest) == c.maxCountTasks {
			return nil, errors.New("Limit on the number of tasks has been reached. ")
		}
		all := len(c.tasksToOneRequest) + len(c.tasksToGroupRequests)
		if all == c.maxCountTasks {
			return nil, errors.New("Limit on the number of tasks has been reached. ")
		}
	}
	return func(request Request, api LevelAPI, signalChannel chan bool, taskStateChannel chan *TaskState) (RunTask, NoWait, int) {
		var runTask = func() {
			c.taskOneRequest(request, api, signalChannel, taskStateChannel)
		}
		if c.countNowExecuteTask == 0 {
			return runTask, true, 0
		}
		if len(c.tasksToOneRequest) != 0 || c.countNowExecuteTask == 1 {
			c.tasksToOneRequest = append(c.tasksToOneRequest, runTask)
		}
		return runTask, false, len(c.tasksToOneRequest) - 1
	}, nil
}

func (c *GithubClient) DropReservedOneRequestTask(index int) {
	c.tasksToOneRequest = append(c.tasksToOneRequest[:index], c.tasksToOneRequest[index+1:]...)
}

// Для создания новой задачи на выполнение группы запросов
// к GitHub, необходимо сначала проверить состояние очереди:
// 	* если она переполнена возвращается ошибка.
// 	* если место в очереди есть, возвращай функцию
// 	  настройки запроса к GitHub (TaskGroupRequests).
//
// Аргументами TaskGroupRequests являются параметры запроса:
// 	* request Request 		  		                  - содержит URL и HEADER запроса.
// 	* api LevelAPI 					                  - уровень API GitHub (Core, Search).
// 	* responsesChannel chan map[string]*Response 	  - канал передачи ответов от API GitHub,
// 														передаются ответы на уже выполненые запросы,
// 														без достижения Rate Limit.
// 	* deferResponsesChannel chan map[string]*Response - канал передачи ответов от API GitHub,
//														передаются ответы на уже выполненые запросы,
//														до момента достижения Rate Limit,
//														а отсальные ответы будут переданы позже,
// 									   					соответсвенно не следует ждать завершения задачи.
func (c *GithubClient) AddGroupRequests(reserved bool) (TaskGroupRequests, error) {
	if !reserved {
		if len(c.tasksToGroupRequests) == c.maxCountTasks {
			return nil, errors.New("Limit on the number of tasks has been reached. ")
		}
		all := len(c.tasksToOneRequest) + len(c.tasksToGroupRequests)
		if all == c.maxCountTasks {
			return nil, errors.New("Limit on the number of tasks has been reached. ")
		}
	}
	return func(requests []Request, api LevelAPI, taskStateChannel, deferTaskStateChannel chan *TaskState) (RunTask, NoWait, int) {
		var runTask = func() {
			c.taskGroupRequests(requests, api, taskStateChannel, deferTaskStateChannel)
		}
		if c.countNowExecuteTask == 0 {
			return runTask, true, 0
		}
		if len(c.tasksToGroupRequests) != 0 || c.countNowExecuteTask == 1 {
			c.tasksToGroupRequests = append(c.tasksToGroupRequests, runTask)
		}
		return runTask, false, len(c.tasksToGroupRequests) - 1
	}, nil
}

func (c *GithubClient) DropReservedGroupRequestTask(index int) {
	c.tasksToGroupRequests = append(c.tasksToGroupRequests[:index], c.tasksToGroupRequests[index+1:]...)
}