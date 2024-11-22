package main

import (
	"fmt"
)

// Stage представляет этап обработки с его параметрами.
type Stage struct {
	Name        string      `yaml:"name"`         // Имя этапа
	Description string      `yaml:"desc"`         // Описание этапа
	Dependence  interface{} `yaml:"dependence"`   // Зависимости этапа
	Atomic      bool        `yaml:"atomic"`       // Флаг атомарности: если true, этап останавливается при ошибке
	PreCheck    Check       `yaml:"pre_check"`    // Предварительная проверка перед выполнением этапа
	PreScript   Script      `yaml:"pre_script"`   // Предварительный скрипт перед выполнением этапа
	PostCheck   Check       `yaml:"post_check"`   // Пост-проверка после выполнения этапа
	PostScript  Script      `yaml:"post_scriprt"` // Пост-скрипт после выполнения этапа
	Rollback    bool        `yaml:"rollback"`     // Флаг отката: если true, позволяет откатить изменения
	Steps       []Stage     `yaml:"step"`         // Шаги, которые входят в этот этап
}

var ATOMIC_STAGE bool // Глобальный флаг атомарности текущего этапа

// processStage обрабатывает этап (Stage) с различными проверками и скриптами.
func processStage(stage Stage, parentAtomic bool) error {
	logMessage("INFO", fmt.Sprintf("========== Start stage: %s ==========", stage.Name))

	// Если указано описание этапа, выводим его в лог
	printDescription(stage.Description)

	// Если родительский этап не атомарен, проверяем флаг атомарности текущего этапа
	if !parentAtomic {
		ATOMIC_STAGE = stage.Atomic
		logMessage("INFO", fmt.Sprintf("ATOMIC Stage: %v", ATOMIC_STAGE))
	} else {
		// Если родительский этап атомарен, текущий этап также наследует атомарность
		logMessage("INFO", "ATOMIC Parent")
	}

	// Выполняем предварительные проверки, если они указаны
	if err := checkHandler(stage, "preCheck", ATOMIC_STAGE); err != nil {
		return err
	}

	// Выполняем предварительные действия (скрипты), если они указаны
	if err := actionHandler(stage, "preScript", ATOMIC_STAGE); err != nil {
		return err
	}

	// Последовательно обрабатываем каждый шаг этапа
	for _, step := range stage.Steps {
		if err := processStage(step, ATOMIC_STAGE); err != nil {
			return err
		}
	}

	// Выполняем пост-проверки после завершения этапа
	if err := checkHandler(stage, "postCheck", ATOMIC_STAGE); err != nil {
		return err
	}

	// Выполняем пост-действия (скрипты) после завершения этапа
	if err := actionHandler(stage, "postScript", ATOMIC_STAGE); err != nil {
		return err
	}

	return nil
}
