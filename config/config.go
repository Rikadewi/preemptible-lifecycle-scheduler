package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	EnvProduction  = "production"
	EnvDevelopment = "development"
)

type Config struct {
	Environment    string   `yaml:"environment"`
	PeakHourRanges []string `yaml:"peak-hour-ranges"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Environment:    EnvDevelopment,
		PeakHourRanges: []string{},
	}
}

func (config *Config) Load(filepath string) (err error) {
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return
	}

	return yaml.Unmarshal(yamlFile, config)
}
