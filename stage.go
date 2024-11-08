package main

import "fmt"

type Stage struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"desc"`
	Dependence  interface{} `yaml:"dependence"`
	Atomic      bool        `yaml:"atomic"`
	PreCheck    Check       `yaml:"pre_check"`
	PreScript   Script      `yaml:"pre_script"`
	PostCheck   Check       `yaml:"post_check"`
	PostScript  Script      `yaml:"post_scriprt"`
	Rollback    bool        `yaml:"rollback"`
	Steps       []Step      `yaml:"step"`
}

// Константа для атомарного обновления
var ATOMIC_STAGE bool

func processStage(stage Stage) error {

	// Проверка наличия подписи
	if stage.Description != "" {
		logMessage("INFO", fmt.Sprintf("Desc: %s", stage.Description))
	}

	// Проверка наличия атомарного этапа
	if stage.Atomic {
		ATOMIC_STAGE = stage.Atomic
		logMessage("DEBUG", fmt.Sprintf("ATOMIC_STAGE: %v", ATOMIC_STAGE))
	}

	// Проверка наличия пре-чека
	if (Check{}) != stage.PreCheck {
		logMessage("INFO", fmt.Sprintf("Running pre-check for stage: %s", stage.Name))
		if err := executeCheck(stage.PreCheck); err != nil {
			return fmt.Errorf("pre-check failed for stage %s: %v", stage.Name, err)
		}
	} else {
		logMessage("DEBUG", fmt.Sprintf("pre_check for stage %s is missing", stage.Name))
	}

	// Проверка наличия пре-скрипта
	if (Script{}) != stage.PreScript {
		logMessage("INFO", fmt.Sprintf("Running pre-script for stage: %s", stage.Name))
		if err := executeCheck(stage.PreCheck); err != nil {
			return fmt.Errorf("pre-script failed for stage %s: %v", stage.Name, err)
		}
	} else {
		logMessage("DEBUG", fmt.Sprintf("pre-script for stage %s is missing", stage.Name))
	}

	for _, step := range stage.Steps {
		logMessage("INFO", fmt.Sprintf("Processing step: %s", step.Name))
		if err := processStep(step); err != nil {
			return err
		}
	}

	logMessage("INFO", fmt.Sprintf("Running post-check for stage: %s", stage.Name))
	if err := executeCheck(stage.PostCheck); err != nil {
		return fmt.Errorf("post-check failed for stage %s: %v", stage.Name, err)
	}

	return nil
}
