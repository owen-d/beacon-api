package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

type JsonConfig struct {
	Scope            string `json:"scope"`
	GCloudConfigPath string `json:"gCloudConfigPath"`
	JWTSecret        string `json:"JWTSecret"`
}

func LoadConfFromDir(fPath string) (*JsonConfig, error) {

	conf := &JsonConfig{
		// default gcp config location is in same dir as config.json
		GCloudConfigPath: filepath.Join(fPath, "gcp-credentials.json"),
	}
	data, err := ioutil.ReadFile(filepath.Join(fPath, "config.json"))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, conf)

	return conf, err
}
