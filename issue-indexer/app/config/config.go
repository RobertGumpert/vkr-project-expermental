package config

import (
	"encoding/json"
	"io/ioutil"
	"issue-indexer/pckg/runtimeinfo"
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
	CountTask                 int64    `json:"count_task"`
	GithubCollectorsAddresses []string `json:"github_collectors_addresses"`
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
