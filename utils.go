package main

import (
	"fmt"
	"reflect"
)

// printDescription выводит описание этапа, если оно есть.
func printDescription(description string) {
	if description != "" {
		logMessage("INFO", fmt.Sprintf("========== %s", description))
	}
}

// handleAtomicStage устанавливает флаг атомарного обновления, если он присутствует.
func handleAtomicStage(isAtomic bool) {
	if isAtomic {
		ATOMIC_STAGE = isAtomic
		logMessage("DEBUG", fmt.Sprintf("ATOMIC_STAGE: %v", ATOMIC_STAGE))
	}
}

// runPreActions выполняет pre-check и pre-script, если они есть.
func runPreActions(data interface{}) error {
	logMessage("DEBUG", "Starting pre-actions...")
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	preCheckField := val.FieldByName("pre-check")
	if !preCheckField.IsValid() || preCheckField.Kind() != reflect.Struct {
		logMessage("DEBUG", "Missing Pre-check")
	} else {
		if err := runAction(preCheckField, "pre-check", "PreCheck"); err != nil {
			return err
		}
	}

	locationField := val.FieldByName("pre-script")
	if !locationField.IsValid() || locationField.Kind() != reflect.Struct {
		logMessage("DEBUG", "Missing pre-script")
	} else {
		if err := runAction(locationField, "pre-script", "preScript"); err != nil {
			return err
		}
	}

	return nil
}

// runPostActions выполняет post-check и post-script, если они есть.
func runPostActions(stage Stage) error {
	logMessage("DEBUG", "Starting post-actions...")
	if err := runAction(stage.PostScript, "post-script", "PostScript"); err != nil {
		return err
	}
	return runAction(stage.PostCheck, "post-check", "PostCheck")
}

// runAction выполняет проверку или скрипт, если они присутствуют.
func runAction(action interface{}, actionName, actionType string) error {
	switch v := action.(type) {
	case Check:
		// Если это Check, выполняем проверку
		logMessage("INFO", fmt.Sprintf("Running %s for stage", actionName))
		if err := executeCheck(v); err != nil {
			return fmt.Errorf("%s failed for stage: %v", actionType, err)
		}
	case Script:
		// Если это Script, выполняем скрипт
		logMessage("INFO", fmt.Sprintf("Running %s for stage", actionName))
		if err := executeScript(v); err != nil {
			return fmt.Errorf("%s failed for stage: %v", actionType, err)
		}
	default:
		// Если это не Check или Script, выводим debug сообщение
		logMessage("DEBUG", fmt.Sprintf("%s for stage is missing", actionType))
	}
	return nil
}
