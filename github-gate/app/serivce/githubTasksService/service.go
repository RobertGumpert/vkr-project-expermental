package githubTasksService

import (
	"github-gate/app/config"
	"github-gate/pckg/runtimeinfo"
	"github-gate/pckg/task"
	"net/http"
)

type QueueIsBusy func() bool
type SendTaskToGithubCollector func()

type GithubTasksService struct {
	config                  *config.Config
	client                  *http.Client
	tasksForCollectorsQueue []*TaskForCollector
	completedTasksChannel   chan *TaskForCollector
}

func NewGithubTasksService(config *config.Config, client *http.Client) *GithubTasksService {
	tasksForCollectorsQueue := make([]*TaskForCollector, 0)
	tasksForCollectorsChannel := make(chan *TaskForCollector, config.SizeQueueTasksForGithubCollectors)
	service := &GithubTasksService{
		config:                  config,
		client:                  client,
		tasksForCollectorsQueue: tasksForCollectorsQueue,
		completedTasksChannel:   tasksForCollectorsChannel,
	}
	go service.scanCompletedTasksChannel()
	return service
}

func (service *GithubTasksService) CreateTaskRepositoriesDescriptions(taskFromTaskService task.ITask, repositoriesUrls []string) (QueueIsBusy, SendTaskToGithubCollector) {
	var (
		queueHasPlace = true
		queueIsBusy   = func() bool {
			queueHasPlace = service.queueHasFreeSpace(1)
			return queueHasPlace
		}
		sendTaskToGithubCollector = func() {
			if !queueHasPlace {
				return
			}
			taskForCollector, isDefer := service.createTaskRepositoriesDescriptionByURL(
				taskFromTaskService,
				repositoriesUrls,
			)
			service.tasksForCollectorsQueue = append(
				service.tasksForCollectorsQueue,
				taskForCollector,
			)
			if !isDefer {
				err := service.pipelineSendTaskToCollector(taskForCollector)
				if err != nil {
					runtimeinfo.LogError("SEND NEW TASK TO COLLECTOR WAS COMPLETED WITH ERROR: ", err)
				}
			}
			runtimeinfo.LogInfo("ADD NEW TASK [", taskForCollector.GetKey(), "]")
		}
	)
	return queueIsBusy, sendTaskToGithubCollector
}

func (service *GithubTasksService) CreateTaskRepositoriesAndTheirIssues(taskFromTaskService task.ITask, repositoriesUrls []string) (QueueIsBusy, SendTaskToGithubCollector) {
	var (
		queueHasPlace = true
		queueIsBusy   = func() bool {
			countTasks := 1 + len(repositoriesUrls)
			queueHasPlace = service.queueHasFreeSpace(countTasks)
			return queueHasPlace
		}
		sendTaskToGithubCollector = func() {
			if !queueHasPlace {
				return
			}
			trigger, isDefer := service.createTaskRepositoriesDescriptionByURL(
				taskFromTaskService,
				repositoriesUrls,
			)
			dependent, _ := service.createTasksListRepositoriesIssues(
				taskFromTaskService,
				repositoriesUrls,
			)
			service.linkTriggerWithDependentTasks(trigger, dependent)
			service.tasksForCollectorsQueue = append(
				service.tasksForCollectorsQueue,
				trigger,
			)
			service.tasksForCollectorsQueue = append(
				service.tasksForCollectorsQueue,
				dependent...,
			)
			var dependsKey string
			for i := 0; i < len(dependent); i++ {
				dependsKey += dependent[i].GetKey() + ", "
			}
			runtimeinfo.LogInfo("ADD NEW TRIGGER TASK [", trigger.GetKey(), "]")
			runtimeinfo.LogInfo("ADD NEW DEPENDENT TASKS [", dependsKey, "]")
			if !isDefer {
				err := service.pipelineSendTaskToCollector(trigger)
				if err != nil {
					runtimeinfo.LogError("SEND NEW TASK TO COLLECTOR WAS COMPLETED WITH ERROR: ", err)
				}
			}
		}
	)
	return queueIsBusy, sendTaskToGithubCollector
}

func (service *GithubTasksService) scanCompletedTasksChannel() {
	for taskForCollector := range service.completedTasksChannel {
		switch taskForCollector.GetType() {
		case taskTypeRepositoriesDescriptionsByURL:
			runtimeinfo.LogInfo("TASK [", taskForCollector.GetKey(), "] WAS COMPLETED.")
			break
		case taskTypeRepositoriesDescriptionsAndTheirIssues:
			if taskForCollector.details.IsTrigger() {
				if existDependentTasks, _ := taskForCollector.details.HasDependentTasks(); !existDependentTasks {
					runtimeinfo.LogError("COMPLETED TRIGGER TASK [", taskForCollector.GetKey(), "] DOESN'T HAVE DEPENDENT TASKS.")
					break
				}
				err := service.sendToCollectorsDependTasks(taskForCollector)
				if err != nil {
					runtimeinfo.LogError("COMPLETED TRIGGER TASK [", taskForCollector.GetKey(), "] WAS COMPLETED WITH ERROR: ", err)
					break
				}
				runtimeinfo.LogInfo("COMPLETED TRIGGER TASK [", taskForCollector.GetKey(), "] RUN DEPENDENT TASKS.")
			}
			if taskForCollector.details.IsDependent() {
				trigger := taskForCollector.details.GetTriggerTask()
				if trigger == nil {
					runtimeinfo.LogError("COMPLETED DEPEND TASK [", taskForCollector.GetKey(), "] DOESN'T HAVE TRIGGER.")
					break
				}
				isTrigger, dependTasks := trigger.details.HasDependentTasks()
				if isTrigger == false || len(dependTasks) == 0 {
					runtimeinfo.LogError("TRIGGER WITH A COMPLETED, DEPENDENT TASK [", taskForCollector.GetKey(), "] DOESN'T HAVE DEPENDENT TASKS.")
					break
				}
				countCompletedTasks := trigger.details.CountCompletedDependentTasks()
				runtimeinfo.LogInfo("COMPLETED DEPEND DEPEND TASK [", taskForCollector.GetKey(), "] RUN BY TRIGGER TASK [", trigger.GetKey(), "] WAS COMPLETED.")
				if countCompletedTasks == len(dependTasks) {
					runtimeinfo.LogInfo("TRIGGER TASK [", trigger.GetKey(), "] WAS COMPLETED.")
				}
				break
			}
		}
		service.sendToCollectorsDeferTasks()
	}
}
