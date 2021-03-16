package githubTasksService

import (
	"fmt"
	"github-gate/app/models/sendTaskModel"
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
func (service *GithubTasksService) getTaskRepositoriesDescriptionByURL(repositoriesUrls []string) (*TaskForCollector, bool) {
	for url := 0; url < len(repositoriesUrls); url++ {
		if !strings.Contains(repositoriesUrls[url], gitHubApiAddress) {
			repositoriesUrls[url] = strings.Join([]string{gitHubApiAddress, "repos", repositoriesUrls[url]}, "/")
		}
	}
	var (
		isDeferTask      = false
		taskForCollector = NewTaskForCollector(
			TaskTypeRepositoriesDescriptionsByURL,
			"",
			false,
			false,
			false,
			nil,
			new(TaskDetails),
		)
	)
	isDeferTask = service.setCollectorForTask(
		taskForCollector,
		collectorEndpointRepositoriesByURL,
	)
	taskForCollector.taskDetails.sendToCollectorJsonBody = &sendTaskModel.RepositoriesByURLS{
		TaskKey: taskForCollector.key,
		URLS:    repositoriesUrls,
	}
	return taskForCollector, isDeferTask
}

// INPUT:
// 			-> [https://api.github.com/<user>/<name>, ..., n]
// 			or
// 			-> [<user>/<name>, ..., n]
//
func (service *GithubTasksService) getListTasksRepositoriesIssues(repositoriesUrls []string) ([]*TaskForCollector, concurrentMap.ConcurrentMap) {
	var (
		deferTasks = concurrentMap.New()
		tasks      = make([]*TaskForCollector, 0)
	)
	for url := 0; url < len(repositoriesUrls); url++ {
		if !strings.Contains(repositoriesUrls[url], gitHubApiAddress) {
			repositoriesUrls[url] = strings.Join([]string{gitHubApiAddress, "repos", repositoriesUrls[url]}, "/")
		}
		taskForCollector := NewTaskForCollector(
			TaskTypeRepositoryIssues,
			"",
			false,
			false,
			false,
			nil,
			new(TaskDetails),
		)
		if isDeferTask := service.setCollectorForTask(
			taskForCollector,
			collectorEndpointRepositoryIssues,
		); isDeferTask {
			deferTasks.Set(taskForCollector.key, taskForCollector)
		}
		tasks = append(
			tasks,
			taskForCollector,
		)
	}
	return tasks, deferTasks
}

func (service *GithubTasksService) linkDependentTasks(initializer *TaskForCollector, dependent []*TaskForCollector) {
	for i := 0; i < len(dependent); i++ {
		dependent[i].taskDetails.signalTriggeredDependentTask = true
	}
	initializer.taskDetails.dependentTasksRunAfterCompletion = dependent
}

func (service *GithubTasksService) setCollectorForTask(taskForCollector *TaskForCollector, collectorEndpoint string) bool {
	var (
		isSetCollector          = false
		taskKey                 = ""
		taskNumber              = service.tasksForCollectors.Count() + 1
		freeCollectorsAddresses = service.getFreeCollectors()
	)
	if freeCollectorsAddresses == nil {
		taskKey = fmt.Sprintf("defer-task-key-[%d]", taskNumber)
		taskForCollector.deferStatus = true
	} else {
		taskKey = fmt.Sprintf("task-key-[%d]", taskNumber)
		freeCollectorAddress := freeCollectorsAddresses[0]
		taskForCollector.taskDetails.collectorAddress = freeCollectorAddress
		taskForCollector.taskDetails.collectorURL = fmt.Sprintf(
			"%s/%s",
			freeCollectorAddress,
			collectorEndpoint,
		)
		isSetCollector = true
	}
	if strings.TrimSpace(taskForCollector.key) == "" {
		taskForCollector.key = taskKey
	}
	return isSetCollector
}
