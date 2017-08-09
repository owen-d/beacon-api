package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type JsonConfig struct {
	Scope            string `json:"scope"`
	GCloudConfigPath string `json:"gCloudConfigPath"`
	JWTSecret        string `json:"JWTSecret"`
	CassEndpoint     string
	CassKeyspace     string
	Port             int
}

func LoadConfFromDir(fPath string) (*JsonConfig, error) {

	cassEndpoint := os.Getenv("CASSANDRA_ENDPOINT")
	if cassEndpoint == "" {
		cassEndpoint = "localhost"
	}
	var port int
	if strPort := os.Getenv("LISTEN_PORT"); strPort == "" {
		port = 8080
	} else {
		parsedPort, parseErr := strconv.Atoi(strPort)
		if parseErr != nil {
			return nil, parseErr
		}
		port = parsedPort
	}
	conf := &JsonConfig{
		// default gcp config location is in same dir as config.json
		GCloudConfigPath: filepath.Join(fPath, "gcp-credentials.json"),
		CassEndpoint:     cassEndpoint,
		// hardcode keyspace
		CassKeyspace: "bkn",
		// hardcoded port
		Port: port,
	}
	data, err := ioutil.ReadFile(filepath.Join(fPath, "config.json"))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, conf)

	return conf, err
}
