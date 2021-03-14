package createTaskModel

type CreateTaskCompareIssuesInPairs struct {
	TaskKey                   string `json:"task_key"`
	ComparableRepositoryID    uint   `json:"comparable_repository_id"`
	CompareWithRepositoriesID []uint `json:"compare_with_repositories_id"`
}
