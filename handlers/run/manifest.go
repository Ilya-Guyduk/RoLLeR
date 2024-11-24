package run

import "fmt"

var ATOMIC_MANIFEST string

func manifestHandler(manifestPath string) error {

	// Валидация YML манифеста
	config, err := validateYAML(manifestPath)
	if err != nil {
		logMessage("ERROR", fmt.Sprintf("Error validating YAML file: %v", err))
		return err
	}
	if config.Version != "" {
		logMessage("INFO", "Release version: %s", config.Version)
	}

	if config.Atomic == nil {
		logMessage("DEBUG", "MANIFEST Atomic flag is not specified")
	} else if *config.Atomic {
		logMessage("DEBUG", "MANIFEST Atomic is true")
	} else {
		logMessage("DEBUG", "MANIFEST Atomic is false")
	}

	if *&DRY_RUN_FLAG {

		for _, stage := range config.Stages {
			if err := processStage(stage, config.Atomic); err != nil {
				logMessage("ERROR", fmt.Sprintf("Error processing stage %s: %v", stage.Name, err))
			}
		}
		return nil
	}

	// Обработка этапов в манифесте
	for _, stage := range config.Stages {
		if err := processStage(stage, config.Atomic); err != nil {
			if *config.Atomic {
				return err
			}
		}
	}

	return nil
}
