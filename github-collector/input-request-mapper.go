package main

type CreateTaskReposByURL struct {
	TaskKey string   `json:"task_key"`
	URLS    []string `json:"urls"`
}

type CreateTaskRepositoryIssues struct {
	TaskKey string `json:"task_key"`
	URL     string `json:"url"`
}
