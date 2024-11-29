package run

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Atomic  *bool   `yaml:"atomic"`
	Version string  `yaml:"version"`
	Stages  []Stage `yaml:"stage"`
}

func validateYAML(configPath string) (*Manifest, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	var manifest Manifest
	err = yaml.Unmarshal(data, &manifest)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	if len(manifest.Stages) == 0 {
		return nil, fmt.Errorf("no stages found in the YAML configuration")
	}

	for _, stage := range manifest.Stages {
		if stage.Name == "" {
			return nil, fmt.Errorf("stage name is missing for stage: %+v", stage)
		}

		if len(stage.Steps) == 0 {
			return nil, fmt.Errorf("no steps found for stage: %s", stage.Name)
		}
	}

	return &manifest, nil
}
