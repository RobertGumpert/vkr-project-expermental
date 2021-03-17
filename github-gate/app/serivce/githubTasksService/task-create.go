package githubTasksService

import (
	"github-gate/app/models/interapplicationModels/githubCollectorModels"
	"github-gate/pckg/task"
	concurrentMap "github.com/streamrail/concurrent-map"
	"strings"
)

const (
	gitHubApiAddress = "https://api.github.com"
	//
	collectorEndpointRepositoriesByURL = "/get/repos/by/url"
	collectorEndpointRepositoryIssues  = "/get/repos/issues"
)

// INPUT:
// 			-> [https://api.github.com/<user>/<name>, ..., n]
// 			or
// 			-> [<user>/<name>, ..., n]
//
func (service *GithubTasksService) createTaskRepositoriesDescriptionByURL(taskFromTaskService task.ITask, repositoriesUrls []string) (*TaskForCollector, bool) {
	for url := 0; url < len(repositoriesUrls); url++ {
		if !strings.Contains(repositoriesUrls[url], gitHubApiAddress) {
			repositoriesUrls[url] = strings.Join(
				[]string{
					gitHubApiAddress,
					"repos",
					repositoriesUrls[url]},
				"/",
			)
		}
	}
	taskForCollector, isDeferTask := service.createTask(
		taskFromTaskService,
		taskTypeRepositoriesDescriptionsByURL,
		collectorEndpointRepositoriesByURL,
	)
	taskForCollector.details.SetSendToCollectorJsonBody(
		&githubCollectorModels.SendTaskRepositoriesByURLS{
			TaskKey: taskForCollector.key,
			URLS:    repositoriesUrls,
		},
	)
	return taskForCollector, isDeferTask
}

// INPUT:
// 			-> [https://api.github.com/<user>/<name>, ..., n]
// 			or
// 			-> [<user>/<name>, ..., n]
//
func (service *GithubTasksService) createTasksListRepositoriesIssues(taskFromTaskService task.ITask, repositoriesUrls []string) ([]*TaskForCollector, concurrentMap.ConcurrentMap) {
	var (
		deferTasks         = concurrentMap.New()
		tasksForCollectors = make([]*TaskForCollector, 0)
	)
	for url := 0; url < len(repositoriesUrls); url++ {
		if !strings.Contains(repositoriesUrls[url], gitHubApiAddress) {
			repositoriesUrls[url] = strings.Join(
				[]string{
					gitHubApiAddress,
					"repos",
					repositoriesUrls[url]},
				"/",
			)
		}
		taskForCollector, isDeferTask := service.createTask(
			taskFromTaskService,
			taskTypeRepositoryIssues,
			collectorEndpointRepositoryIssues,
		)
		if isDeferTask {
			deferTasks.Set(taskForCollector.GetKey(), taskForCollector)
		}
		tasksForCollectors = append(
			tasksForCollectors,
			taskForCollector,
		)
		taskForCollector.details.SetSendToCollectorJsonBody(
			&githubCollectorModels.SendTaskRepositoryIssues{
				TaskKey: taskForCollector.GetKey(),
				URL:     repositoriesUrls[url],
			},
		)
	}
	return tasksForCollectors, deferTasks
}

func (service *GithubTasksService) createTask(taskFromTaskService task.ITask, taskType task.Type, collectorEndpoint string) (*TaskForCollector, bool) {
	var (
		isDeferTask      = false
		taskNumber       = len(service.tasksForCollectorsQueue) + 1
		taskForCollector = newTaskForCollector(
			taskType,
			"",
			false,
			false,
			false,
			nil,
			new(TaskDetails),
		)
	)
	taskForCollector.details.SetTaskFromTaskService(taskFromTaskService)
	taskForCollector.details.SetNumber(taskNumber)
	nonFreeCollectors := service.findAndSetCollectorForNewTask(
		taskForCollector,
		collectorEndpoint,
	)
	if nonFreeCollectors {
		service.createAndSetNewKeyForTask(
			taskForCollector,
			task.DeferType,
		)
		taskForCollector.SetDeferStatus(true)
		isDeferTask = true
	} else {
		service.createAndSetNewKeyForTask(
			taskForCollector,
			task.RunnableType,
		)
	}
	return taskForCollector, isDeferTask
}

func (service *GithubTasksService) linkTriggerWithDependentTasks(trigger *TaskForCollector, dependent []*TaskForCollector) {
	triggerTaskTypes := []task.Type{task.TriggerType}
	if trigger.GetDeferStatus() {
		triggerTaskTypes = append(
			triggerTaskTypes,
			task.DeferType,
		)
	} else {
		triggerTaskTypes = append(
			triggerTaskTypes,
			task.RunnableType,
		)
	}
	service.createAndSetNewKeyForTask(
		trigger,
		triggerTaskTypes...,
	)
	for i := 0; i < len(dependent); i++ {
		dependent[i].SetType(taskTypeRepositoriesDescriptionsAndTheirIssues)
		dependent[i].details.SetTriggerTask(trigger)
		dependent[i].details.SetDependentStatus(true)
		dependent[i].details.SetTriggeredStatus(false)
		dependent[i].SetDeferStatus(true)
		service.createAndSetNewKeyForTask(
			dependent[i],
			task.DependType,
			task.DeferType,
		)
	}
	trigger.SetType(taskTypeRepositoriesDescriptionsAndTheirIssues)
	trigger.details.SetDependentStatus(false)
	trigger.details.SetTriggeredStatus(true)
	trigger.details.SetDependentTasks(dependent)
	return
}

func (service *GithubTasksService) createAndSetNewKeyForTask(taskForCollector *TaskForCollector, taskTypes ...task.Type) {
	taskForCollector.SetKey(
		task.GenerateUniqueKey(
			taskForCollector.details.GetNumber(),
			taskTypes...,
		),
	)
	return
}

func (service *GithubTasksService) swapRunnableAndDeferStatusInKey(taskForCollector *TaskForCollector) {
	taskForCollector.SetKey(
		task.SwapRunnableAndDefer(
			taskForCollector.GetKey(),
		),
	)
}
