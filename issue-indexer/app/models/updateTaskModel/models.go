package updateTaskModel

type UpdateNearestIssues struct {
	DbID             uint `json:"db_id"`
	CompareIssue     uint `json:"compare_issue"`
	NearestWithIssue uint `json:"nearest_with_issue"`
}

type UpdateTaskCompareIssuesInPairs struct {
	TaskKey                   string `json:"task_key"`
	ComparableRepositoryID    uint   `json:"comparable_repository_id"`
	CompareWithRepositoriesID []uint `json:"compare_with_repositories_id"`
	//
	CountNearestIssues int                   `json:"count_nearest_issues"`
	NearestIssues      []UpdateNearestIssues `json:"nearest_issues"`
}
