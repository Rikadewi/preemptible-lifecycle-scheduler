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
	IncludedPool   string   `yaml:"included-pool"`
	ExcludedPool   string   `yaml:"excluded-pool"`
	GracefulPeriod int      `yaml:"graceful-period"`
	PeakHourRanges []string `yaml:"peak-hour-ranges"`
	Debug          bool     `yaml:"debug"`
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
