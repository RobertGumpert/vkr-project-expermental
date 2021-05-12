package config

import (
	"encoding/json"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	Port     string `json:"port"`
	Postgres struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Port     string `json:"port"`
		DbName   string `json:"db_name"`
		Ssl      string `json:"ssl"`
	} `json:"postgres"`
	//
	//
	//
	MaximumSizeOfQueue             int64   `json:"maximum_size_of_queue"`
	MinimumCosineDistanceThreshold float64 `json:"minimum_cosine_distance_threshold"`
	//
	// GITHUB-GATE
	//
	GithubGateAddress   string `json:"github_gate_address"`
	GithubGateEndpoints struct {
		SendResultTaskReindexingForAll               string `json:"send_result_task_reindexing_for_all"`
		SendResultTaskReindexingForRepository        string `json:"send_result_task_reindexing_for_repository"`
		SendResultTaskReindexingForGroupRepositories string `json:"send_result_task_reindexing_for_group_repositories"`
	} `json:"github_gate_endpoints"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Read() *Config {
	absPath, err := filepath.Abs("../repository-indexer/data/config/config.json")
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
