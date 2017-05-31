package config

import (
	"encoding/json"
	"io/ioutil"
)

type JsonConfig struct {
	Scope            string `json:"scope"`
	GCloudConfigPath string `json:"gCloudConfigPath"`
}

func LoadConfFromFile(fPath string) (*JsonConfig, error) {
	var conf JsonConfig
	data, err := ioutil.ReadFile(fPath)
	if err != nil {
		return &conf, err
	}

	err = json.Unmarshal(data, &conf)

	return &conf, err
}
