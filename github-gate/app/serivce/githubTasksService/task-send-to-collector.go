package githubTasksService

import (
	"errors"
	"fmt"
	"github-gate/pckg/requests"
	"github-gate/pckg/runtimeinfo"
	"net/http"
	"strings"
)

func (service *GithubTasksService) sendTaskToCollector(taskForCollector *TaskForCollector) error {
	var err error = nil
	if strings.TrimSpace(taskForCollector.details.GetCollectorEndpoint()) == "" {
		return errors.New("NON SET COLLECTOR FOR TASK ["+taskForCollector.GetKey()+"]. ")
	}
	collectorForTaskIsFree := service.collectorIsFree(
		taskForCollector.details.GetCollectorAddress(),
	)
	if !collectorForTaskIsFree {
		err = errors.New("COLLECTOR " + taskForCollector.details.GetCollectorAddress() + " IS BUSY. ")
	} else {
		err = service.requestToCollector(taskForCollector)
	}
	return err
}

func (service *GithubTasksService) repeatSendTaskToOtherCollectors(taskForCollector *TaskForCollector) error {
	var err error = nil
	freeCollectors := service.getFreeCollectors(true)
	if freeCollectors == nil {
		err = errors.New("ALL COLLECTORS IS BUSY. ")
	} else {

		service.setNewCollectorForTask(taskForCollector, freeCollectors[0])
		err = service.requestToCollector(taskForCollector)
	}
	return err
}

func (service *GithubTasksService) pipelineSendTaskToCollector(taskForCollector *TaskForCollector) error {
	var err error = nil
	err = service.sendTaskToCollector(taskForCollector)
	if err != nil {
		err = service.repeatSendTaskToOtherCollectors(taskForCollector)
	}
	if err == nil {
		taskForCollector.SetRunnableStatus(true)
		if taskForCollector.GetDeferStatus() == true {
			service.swapRunnableAndDeferStatusInKey(taskForCollector)
		}
		taskForCollector.SetDeferStatus(false)
	} else {
		taskForCollector.SetRunnableStatus(false)
		if taskForCollector.GetDeferStatus() == false {
			service.swapRunnableAndDeferStatusInKey(taskForCollector)
		}
		taskForCollector.SetDeferStatus(true)
	}
	return err
}

func (service *GithubTasksService) sendToCollectorsDeferTasks() {
	runtimeinfo.LogInfo("START RUNNING DEFER TASKS TO GITHUB-COLLECTOR...")
	for i := 0; i < len(service.tasksForCollectorsQueue); i++ {
		taskForCollector := service.tasksForCollectorsQueue[i]
		if taskForCollector.details.IsDependent() == false {
			if taskForCollector.GetDeferStatus() == true {
				if taskForCollector.GetRunnableStatus() == false {
					err := service.pipelineSendTaskToCollector(taskForCollector)
					if err != nil {
						runtimeinfo.LogError("REQUEST TO SEND TASK: [", taskForCollector.key, "] TO GITHUB-COLLECTOR COMPETED WITH ERROR: ", err)
					}
				}
			}
		}
	}
	runtimeinfo.LogInfo("FINISH RUNNING DEFER TASKS TO GITHUB-COLLECTOR.")
}

func (service *GithubTasksService) sendToCollectorsDependTasks(triggerTask *TaskForCollector) error {
	runtimeinfo.LogInfo("START RUNNING DEPEND TASKS FOR [", triggerTask.GetKey(), "] TO GITHUB-COLLECTOR.")
	var (
		err                    error = nil
		isTrigger, dependTasks       = triggerTask.details.HasDependentTasks()
	)
	if !isTrigger {
		err = errors.New("TASK ISN'T TRIGGER [" + triggerTask.GetKey() + "] WASN'T COMPLETED.")
	} else {
		if !triggerTask.GetExecutionStatus() {
			err = errors.New("TRIGGER TASK [" + triggerTask.GetKey() + "] WASN'T COMPLETED.")
		} else {
			for i := 0; i < len(dependTasks); i++ {
				taskForCollector := dependTasks[i]
				if taskForCollector.details.IsDependent() == true {
					if taskForCollector.GetExecutionStatus() == false {
						err := service.pipelineSendTaskToCollector(taskForCollector)
						if err != nil {
							runtimeinfo.LogError("SEND DEPEND TASK [", taskForCollector.GetKey(), "] COMPLETED WITH ERR: ", err)
						}
					}
				} else {
					runtimeinfo.LogError("TRIGGER TASK [", triggerTask.GetKey(), "] HAVE NOT DEPEND TASK: [", taskForCollector.GetKey(), "]")
				}
			}
		}
	}
	runtimeinfo.LogInfo("FINISH RUNNING DEPEND TASKS TO GITHUB-COLLECTOR.")
	return err
}

func (service *GithubTasksService) requestToCollector(taskForCollector *TaskForCollector) error {
	response, err := requests.POST(
		service.client,
		taskForCollector.details.GetCollectorURL(),
		nil,
		taskForCollector.details.GetSendToCollectorJsonBody(),
	)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("SEND TASK WITH STATUS: %d", response.StatusCode))
	}
	return nil
}
