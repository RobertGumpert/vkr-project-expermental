package githubCollectorModels

type SendTaskRepositoriesByURLS struct {
	TaskKey *string   `json:"task_key"`
	URLS    []string `json:"urls"`
}

type SendTaskRepositoryIssues struct {
	TaskKey *string `json:"task_key"`
	URL     string `json:"url"`
}