package run

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

func validateYAMLConfig(configPath string) (*RollerConfig, *LoggingConfig, *PluginConfig, error) {
	// Чтение файла
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	// Разбор rollerConfig
	var rollerconfig RollerConfig
	err = yaml.Unmarshal(data, &rollerconfig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error parsing YAML for rollerConfig: %v", err)
	}

	// Извлечение блока Logging
	loggingConfig := rollerconfig.Global.Logging
	pluginConfig := rollerconfig.Global.Plugin
	if err != nil {
		return nil, nil, nil, err
	}

	return &rollerconfig, &loggingConfig, &pluginConfig, nil
}

func validateYAMLManifest(manifestPath string) (*Migration, error) {
	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	var manifest Migration
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

		//if len(stage.Stages) == 0 {
		//	return nil, fmt.Errorf("no steps found for stage: %s", stage.Name)
		//}
	}

	return &manifest, nil
}
