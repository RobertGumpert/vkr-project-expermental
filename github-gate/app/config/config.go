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
	SizeQueueTasksForGithubCollectors int64    `json:"size_queue_tasks_for_github_collectors"`
	GithubCollectorsAddresses         []string `json:"github_collectors_addresses"`
	//
	IssueIndexerAddress   string `json:"issue_indexer_address"`
	IssueIndexerEndpoints struct {
		GetState                    string `json:"get_state"`
		CompareForGroupRepositories string `json:"compare_for_group_repositories"`
	} `json:"issue_indexer_endpoints"`
	//
	RepositoryIndexerAddress   string `json:"repository_indexer_address"`
	RepositoryIndexerEndpoints struct {
		GetState                       string `json:"get_state"`
		ReindexingForGroupRepositories string `json:"reindexing_for_group_repositories"`
		ReindexingForRepository        string `json:"reindexing_for_repository"`
		ReindexingForAll               string `json:"reindexing_for_all"`
	} `json:"repository_indexer_endpoints"`
	//
	AppAddress string `json:"app_address"`
	AppEndpoints struct{
		NearestRepositories string `json:"nearest_repositories"`
	} `json:"app_endpoints"`
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
