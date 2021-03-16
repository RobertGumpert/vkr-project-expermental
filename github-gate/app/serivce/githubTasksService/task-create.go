package githubTasksService

import (
	"github-gate/app/models/interapplicationModels/githubCollectorModels"
	"github-gate/pckg/task/generateTaskKey"
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
func (service *GithubTasksService) newCollectorTaskRepositoriesDescriptionByURL(repositoriesUrls []string) (*TaskForCollector, bool) {
	for url := 0; url < len(repositoriesUrls); url++ {
		if !strings.Contains(repositoriesUrls[url], gitHubApiAddress) {
			repositoriesUrls[url] = strings.Join([]string{gitHubApiAddress, "repos", repositoriesUrls[url]}, "/")
		}
	}
	var (
		taskNumber        = len(service.tasksForCollectors) + 1
		nonFreeCollectors = false
		taskForCollector  = newTaskForCollector(
			taskTypeRepositoriesDescriptionsByURL,
			"",
			false,
			false,
			false,
			nil,
			new(TaskDetails),
		)
	)
	taskForCollector.taskDetails.SetNumber(taskNumber)
	nonFreeCollectors = service.findAndSetCollectorForNewTask(
		taskForCollector,
		collectorEndpointRepositoriesByURL,
	)
	if nonFreeCollectors {
		service.setKeyForCollectorTask(taskForCollector, true, generateTaskKey.DeferBehavior)
		taskForCollector.SetDeferStatus(true)
		taskForCollector.SetRunnableStatus(false)
	} else {
		service.setKeyForCollectorTask(taskForCollector, true, generateTaskKey.RunnableBehavior)
		taskForCollector.SetDeferStatus(false)
		taskForCollector.SetRunnableStatus(false)
	}
	taskForCollector.taskDetails.sendToCollectorJsonBody = &githubCollectorModels.SendTaskRepositoriesByURLS{
		TaskKey: taskForCollector.key,
		URLS:    repositoriesUrls,
	}
	return taskForCollector, nonFreeCollectors
}

// INPUT:
// 			-> [https://api.github.com/<user>/<name>, ..., n]
// 			or
// 			-> [<user>/<name>, ..., n]
//
func (service *GithubTasksService) newListCollectorTasksRepositoriesIssues(repositoriesUrls []string) ([]*TaskForCollector, concurrentMap.ConcurrentMap) {
	var (
		deferTasks = concurrentMap.New()
		tasks      = make([]*TaskForCollector, 0)
	)
	for url := 0; url < len(repositoriesUrls); url++ {
		var taskNumber = len(service.tasksForCollectors) + 1
		if !strings.Contains(repositoriesUrls[url], gitHubApiAddress) {
			repositoriesUrls[url] = strings.Join([]string{gitHubApiAddress, "repos", repositoriesUrls[url]}, "/")
		}
		taskForCollector := newTaskForCollector(
			taskTypeRepositoryIssues,
			"",
			false,
			false,
			false,
			nil,
			new(TaskDetails),
		)
		taskForCollector.taskDetails.SetNumber(taskNumber)
		nonFreeCollectors := service.findAndSetCollectorForNewTask(
			taskForCollector,
			collectorEndpointRepositoryIssues,
		)
		if nonFreeCollectors {
			service.setKeyForCollectorTask(taskForCollector, true, generateTaskKey.DeferBehavior)
			taskForCollector.SetDeferStatus(true)
			taskForCollector.SetRunnableStatus(false)
			deferTasks.Set(taskForCollector.key, taskForCollector)
		} else {
			service.setKeyForCollectorTask(taskForCollector, true, generateTaskKey.RunnableBehavior)
			taskForCollector.SetDeferStatus(false)
			taskForCollector.SetRunnableStatus(false)
		}
		tasks = append(
			tasks,
			taskForCollector,
		)
	}
	return tasks, deferTasks
}

func (service *GithubTasksService) linkDependentCollectorTasks(initializer *TaskForCollector, dependent []*TaskForCollector) {
	initializerBehaviors := []generateTaskKey.ExecutionBehavior{generateTaskKey.TriggeredBehavior}
	if initializer.GetDeferStatus() {
		initializerBehaviors = append(
			initializerBehaviors,
			generateTaskKey.DeferBehavior,
		)
	} else {
		initializerBehaviors = append(
			initializerBehaviors,
			generateTaskKey.RunnableBehavior,
		)
	}
	service.setKeyForCollectorTask(
		initializer,
		true,
		initializerBehaviors...,
	)
	for i := 0; i < len(dependent); i++ {
		dependent[i].taskDetails.signalTriggeredDependentTask = true
		dependent[i].deferStatus = true
		service.setKeyForCollectorTask(
			dependent[i],
			true,
			generateTaskKey.DependBehavior,
			generateTaskKey.DeferBehavior,
		)
	}
	initializer.taskDetails.dependentTasksRunAfterCompletion = dependent
	return
}

func (service *GithubTasksService) setKeyForCollectorTask(taskForCollector *TaskForCollector, newKey bool, behavior ...generateTaskKey.ExecutionBehavior) {
	if newKey || strings.TrimSpace(taskForCollector.key) == "" {
		taskForCollector.key = generateTaskKey.GenerateUniqueKey(
			taskForCollector.taskDetails.number,
			behavior...,
		)
	} else {
		taskForCollector.key = generateTaskKey.AddExecutionBehavior(
			taskForCollector.GetKey(),
			behavior...,
		)
	}
	return
}
