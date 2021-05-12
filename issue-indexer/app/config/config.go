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
	MaxCountThreads                       int     `json:"max_count_threads"`
	MinimumTextCompletenessThreshold      float64 `json:"minimum_text_completeness_threshold"`
	MaximumDurationDatabaseQueryInMinutes int     `json:"maximum_duration_database_query_in_minutes"`
	//
	// GITHUB-GATE
	//
	GithubGateAddress   string `json:"github_gate_address"`
	GithubGateEndpoints struct {
		SendResultTaskCompareGroup  string `json:"send_result_task_compare_group"`
		SendResultTaskCompareBeside string `json:"send_result_task_compare_beside"`
	} `json:"github_gate_endpoints"`
}

func NewConfig() *Config {
	return &Config{}
}


func (c *Config) Read() *Config {
	absPath, err := filepath.Abs("../issue-indexer/data/config/config.json")
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
