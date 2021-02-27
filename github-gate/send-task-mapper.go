package main

type SendTaskReposByURL struct {
	TaskKey string   `json:"task_key"`
	URLS    []string `json:"urls"`
}
