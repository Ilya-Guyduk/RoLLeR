package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string  `yaml:"version"`
	Stages  []Stage `yaml:"stage"`
}

func validateYAML(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	if len(config.Stages) == 0 {
		return nil, fmt.Errorf("no stages found in the YAML configuration")
	}

	for _, stage := range config.Stages {
		if stage.Name == "" {
			return nil, fmt.Errorf("stage name is missing for stage: %+v", stage)
		}

		if len(stage.Steps) == 0 {
			return nil, fmt.Errorf("no steps found for stage: %s", stage.Name)
		}
	}

	return &config, nil
}
