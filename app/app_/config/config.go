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
	GithubGateAddress   string `json:"github_gate_address"`
	GithubGateEndpoints struct {
		NewRepositoryNewKeyword      string `json:"new_repository_new_keyword"`
		NewRepositoryExistKeyword    string `json:"new_repository_exist_keyword"`
		ExistRepositoryNewKeyword    string `json:"exist_repository_new_keyword"`
		ExistRepositoryUpdateNearest string `json:"exist_repository_update_nearest"`
	} `json:"github_gate_endpoints"`
	//
	RepositoryIndexerAddress   string `json:"repository_indexer_address"`
	RepositoryIndexerEndpoints struct {
		WordIsExist            string `json:"word_is_exist"`
		GetNearestRepositories string `json:"get_nearest_repositories"`
	} `json:"repository_indexer_endpoints"`
	//
	Posts []struct {
		Host    string `json:"host"`
		TCPPort int    `json:"tcp_port"`
		TLSPort int    `json:"tls_port"`
		Boxes   []struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Identity string `json:"identity"`
		} `json:"boxes"`
	} `json:"posts"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Read() *Config {
	absPath, err := filepath.Abs("../app/data/config/config.json")
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
