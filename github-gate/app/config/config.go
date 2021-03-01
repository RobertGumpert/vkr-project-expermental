package config

import (
	"encoding/json"
	"github-gate/pckg/runtimeinfo"
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	Port                      string   `json:"port"`
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
