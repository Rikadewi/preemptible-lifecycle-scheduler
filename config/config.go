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
	Environment    string   `yaml:"listen-address"`
	PeakHourRanges []string `yaml:"peak-hour-ranges"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Environment:    EnvProduction,
		PeakHourRanges: []string{"03:00-23:00"},
	}
}

func (config *Config) Load(filepath string) (err error) {
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return
	}

	return yaml.Unmarshal(yamlFile, config)
}
