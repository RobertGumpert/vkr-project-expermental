package config

import (
	"encoding/json"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	//
	// DB
	//
	Port     string `json:"port"`
	Postgres struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Port     string `json:"port"`
		DbName   string `json:"db_name"`
		Ssl      string `json:"ssl"`
	} `json:"postgres"`
	//
	// SETTINGS TASK-SERVICE
	//
	MaxCountRunnableTasks int `json:"max_count_runnable_tasks"`
	//
	// SETTINGS COMPARATOR
	//
	MaxCountThreads                  int     `json:"max_count_threads"`
	MinimumTextCompletenessThreshold float64 `json:"minimum_text_completeness_threshold"`
	//
	// GITHUB-GATE
	//
	GithubGateAddress   string `json:"github_gate_address"`
	GithubGateEndpoints struct {
		SendResultTaskCompareIssuesInPairs string `json:"send_result_task_compare_issues_in_pairs"`
	} `json:"github_gate_endpoints"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Read() *Config {
	absPath, err := filepath.Abs("../github-gate/data/config/config.json")
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
