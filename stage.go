package main

import (
	"fmt"
)

// Stage представляет этап обработки с его параметрами.
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

var ATOMIC_STAGE bool

// processStage обрабатывает этап (Stage) с различными проверками и скриптами.
func processStage(stage Stage) error {
	logMessage("INFO", fmt.Sprintf("========== Start stage: %s ==========", stage.Name))

	// Выводим описание, если оно есть
	printDescription(stage.Description)

	// Обрабатываем атомарный флаг
	handleAtomicStage(stage.Atomic)

	/*
		Выполнение предварительные действия, если они указаны
		Запускается функция runPreActions из utils.go
		с переданной структурой Stage
	*/
	if err := runPreActions(stage); err != nil {
		return err
	}

	// Обрабатываем шаги
	for _, step := range stage.Steps {
		if err := processStep(step); err != nil {
			return err
		}
	}

	// Выполняем пост-действия
	if err := runPostActions(stage); err != nil {
		return err
	}

	return nil
}
