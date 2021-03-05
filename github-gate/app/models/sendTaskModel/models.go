package sendTaskModel

type RepositoriesByURLS struct {
	TaskKey string   `json:"task_key"`
	URLS    []string `json:"urls"`
}

type RepositoryIssues struct {
	TaskKey string `json:"task_key"`
	URL     string `json:"url"`
}