package main

import (
	"encoding/json"
	"github-collector/pckg/runtimeinfo"
	"io/ioutil"
	"path/filepath"
)

type config struct {
	Port                string `json:"port"`
	GithubToken         string `json:"github_token"`
	CountTasks          int    `json:"count_tasks"`
	GithubGateAddress   string `json:"github_gate_address"`
	GithubGateEndpoints struct {
		SendResultTaskReposByUlr string `json:"send_result_task_repos_by_ulr"`
		SendResultTaskIssueRepo  string `json:"send_result_task_issue_repo"`
	} `json:"github_gate_endpoints"`
}

func NewConfig() *config {
	return &config{}
}

func (c *config) Read() *config {
	absPath, err := filepath.Abs("../github-collector/data/config/config.json")
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	err = json.Unmarshal(content, c)
	if err != nil {
		runtimeinfo.LogFatal(err)
	}
	return c
}
