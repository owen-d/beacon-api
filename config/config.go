package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type JsonConfig struct {
	Scope            string `json:"scope"`
	GCloudConfigPath string `json:"gCloudConfigPath"`
	JWTSecret        string `json:"JWTSecret"`
	CassEndpoint     string
	CassKeyspace     string
}

func LoadConfFromDir(fPath string) (*JsonConfig, error) {

	cassEndpoint := os.Getenv("CASSANDRA_ENDPOINT")
	if cassEndpoint == "" {
		cassEndpoint = "localhost"
	}
	conf := &JsonConfig{
		// default gcp config location is in same dir as config.json
		GCloudConfigPath: filepath.Join(fPath, "gcp-credentials.json"),
		CassEndpoint:     cassEndpoint,
		// hardcode keyspace
		CassKeyspace: "bkn",
	}
	data, err := ioutil.ReadFile(filepath.Join(fPath, "config.json"))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, conf)

	return conf, err
}
