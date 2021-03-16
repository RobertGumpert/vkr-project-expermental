package tasksService

import (
	"issue-indexer/app/models/dataModel"
	"issue-indexer/app/models/updateTaskModel"
	"issue-indexer/pckg/runtimeinfo"
)

func (service *TasksService) updateTaskCompareIssuesInPairs(taskState *Task) (bool, error) {
	if taskState.GetExecutionStatus() {
		service.countNowRunnableTasks--
		return true, nil
	}
	result := taskState.GetResult().(dataModel.NearestIssues)
	err := service.db.AddNearestIssues(result)
	if err == nil {
		compareReport := taskState.sendResultToGate.(*updateTaskModel.UpdateTaskCompareIssuesInPairs)
		compareReport.CountNearestIssues++
		nearestIssues := updateTaskModel.UpdateNearestIssues{
			DbID:             result.ID,
			CompareIssue:     result.IssueID,
			NearestWithIssue: result.NearestIssueID,
		}
		compareReport.NearestIssues = append(
			compareReport.NearestIssues,
			nearestIssues,
		)
		runtimeinfo.LogInfo("NEAREST ISSUES: key: ", taskState.Key, ", pair:[", result.IssueID, ":", result.NearestIssueID, "]")
	}
	if err != nil {
		runtimeinfo.LogError("NON WRITE TASK STATE: key: ", taskState.GetKey(), ", error: ", err)
	}
	return false, err
}
