package config

import (
	"encoding/json"
	"github-collector/pckg/runtimeinfo"
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	Port                string `json:"port"`
	GithubToken         string `json:"github_token"`
	CountTasks          int    `json:"count_tasks"`
	GithubGateAddress   string `json:"github_gate_address"`
	GithubGateEndpoints struct {
		SendResponseTaskRepositoriesDescriptions string `json:"send_response_task_repositories_descriptions"`
		SendResponseTaskRepositoryIssues         string `json:"send_response_task_repository_issues"`
		SendResponseTaskRepositoriesByKeyWord    string `json:"send_response_task_repositories_by_key_word"`
	} `json:"github_gate_endpoints"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Read() *Config {
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

func (c *Config) ReadWithPath(absPath string) *Config {
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
