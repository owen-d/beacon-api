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
	GoogleOAuth      OAuth `json:"googleOAuth`
}

type OAuth struct {
	ClientID     string `json:"client_id`
	ClientSecret string `json:"client_secret"`
	RedirectUris struct {
		Dev  string `json:"dev"`
		Prod string `json:"prod"`
	} `json:"redirect_uris"`
	Scopes      []string `json:"scopes"`
	StateCipher string
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
	// default configs
	conf := &JsonConfig{
		GCloudConfigPath: filepath.Join(fPath, "gcp-credentials.json"),
		CassEndpoint:     cassEndpoint,
		CassKeyspace:     "bkn",
		Port:             port,
	}
	data, err := ioutil.ReadFile(filepath.Join(fPath, "config.json"))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, conf)

	return conf, err
}
