package config

import (
	"io/ioutil"
	"encoding/json"
)

type LogLevel string
const (
	NoLogging LogLevel = "NoLogging"
	BriefLogging = "BriefLogging"
	FullLogging = "FullLogging"
)

type Config struct {
	Logging struct {
		Requests LogLevel                  `json:"Requests"`
	}                                      `json:"Logging"`
}

func Load(path string, config *Config) (err error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(file, config)
	return
}