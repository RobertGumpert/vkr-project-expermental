package githubTasksService

import (
	"errors"
	"fmt"
	"github-gate/pckg/requests"
	"github-gate/pckg/runtimeinfo"
	"net/http"
)

func (service *GithubTasksService) sendTaskToCollector(taskForCollector *TaskForCollector) (err error, nonFreeCollectors bool) {
	collectorIsFree := service.collectorIsFree(taskForCollector.taskDetails.collectorAddress)
	if !collectorIsFree {
		freeCollectors := service.getFreeCollectors(true)
		if freeCollectors == nil {
			return errors.New("ALL COLLECTORS IS BUSY. "), true
		}
		service.setNewCollectorForTask(taskForCollector, freeCollectors[0])
	}
	err = service.doRequestToCollector(taskForCollector)
	if err != nil {
		taskForCollector.SetDeferStatus(false)
		taskForCollector.SetRunnableStatus(true)
	} else {
		taskForCollector.SetDeferStatus(true)
		taskForCollector.SetRunnableStatus(false)
	}
	return err, false
}

func (service *GithubTasksService) sendDeferTasksToCollectors() {
	runtimeinfo.LogInfo("START RUNNING DEFER TASKS TO GITHUB-COLLECTOR...")
	for i := 0; i < len(service.tasksForCollectors); i++ {
		taskForCollector := service.tasksForCollectors[i]
		if !taskForCollector.GetDeferStatus() && !taskForCollector.GetExecutionStatus() {
			if taskForCollector.GetDeferStatus() && !taskForCollector.taskDetails.IsDependent() {
				err, nonFreeCollectors := service.sendTaskToCollector(taskForCollector)
				if nonFreeCollectors {
					runtimeinfo.LogInfo("FINISH RUNNING DEFER TASKS TO GITHUB-COLLECTOR, BECAUSE ALL COLLECTORS IS BUSY.")
					break
				}
				if err != nil {
					runtimeinfo.LogError("REQUEST TO SEND TASK: [", taskForCollector.key, "] TO GITHUB-COLLECTOR COMPETED WITH ERROR: ", err)
				}
			}
		}
	}
	runtimeinfo.LogInfo("FINISH RUNNING DEFER TASKS TO GITHUB-COLLECTOR.")
}

func (service *GithubTasksService) sendDependTasksToCollectors(triggeredTask *TaskForCollector) error {
	runtimeinfo.LogInfo("START RUNNING DEPEND TASKS FOR [", triggeredTask.GetKey(), "] TO GITHUB-COLLECTOR.")
	var(
		err error = nil
		dependTasks = triggeredTask.taskDetails.dependentTasksRunAfterCompletion
	)
	if !triggeredTask.GetExecutionStatus() {
		err = errors.New("TRIGGERED TASK [" + triggeredTask.GetKey() + "] WASN'T COMPLETED.")
	} else {
		for i := 0; i < len(dependTasks); i++ {
			taskForCollector := service.tasksForCollectors[i]
			err, nonFreeCollectors := service.sendTaskToCollector(taskForCollector)
			if nonFreeCollectors {
				runtimeinfo.LogError("FINISH RUNNING DEPEND TASKS TO GITHUB-COLLECTOR, BECAUSE ALL COLLECTORS IS BUSY.")
				break
			}
			if err != nil {
				runtimeinfo.LogError("SEND DEPEND TASK [", taskForCollector.GetKey(), "] COMPLETED WITH ERR: ", err)
			}
		}
	}
	runtimeinfo.LogInfo("FINISH RUNNING DEPEND TASKS TO GITHUB-COLLECTOR.")
	return err
}

func (service *GithubTasksService) doRequestToCollector(taskForCollector *TaskForCollector) error {
	response, err := requests.POST(
		service.client,
		taskForCollector.taskDetails.GetCollectorURL(),
		nil,
		taskForCollector.taskDetails.GetSendToCollectorJsonBody(),
	)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("SEND TASK WITH STATUS: %d", response.StatusCode))
	}
	return nil
}
