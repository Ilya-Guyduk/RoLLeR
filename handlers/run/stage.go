package run

import (
	"fmt"
)

// Stage представляет этап обработки с его параметрами.
type Stage struct {
	Name        string      `yaml:"name"`         // Имя этапа
	Description string      `yaml:"desc"`         // Описание этапа
	Dependence  interface{} `yaml:"dependence"`   // Зависимости этапа
	Atomic      *bool       `yaml:"atomic"`       // Флаг атомарности: если true, этап останавливается при ошибке
	PreCheck    Check       `yaml:"pre_check"`    // Предварительная проверка перед выполнением этапа
	PreScript   Script      `yaml:"pre_script"`   // Предварительный скрипт перед выполнением этапа
	PostCheck   Check       `yaml:"post_check"`   // Пост-проверка после выполнения этапа
	PostScript  Script      `yaml:"post_scriprt"` // Пост-скрипт после выполнения этапа
	Rollback    bool        `yaml:"rollback"`     // Флаг отката: если true, позволяет откатить изменения
	Steps       []Stage     `yaml:"step"`         // Шаги, которые входят в этот этап
}

var ATOMIC_STAGE *bool // Глобальный флаг атомарности текущего этапа

func descriptor(stage Stage) {
	logMessage("INFO", fmt.Sprintf("========== Start stage: %s ==========", stage.Name))

	// Если указано описание этапа, выводим его в лог
	if stage.Description != "" {
		logMessage("INFO", fmt.Sprintf("========== %s", stage.Description))
	}
}

// processStage обрабатывает этап (Stage) с различными проверками и скриптами.
func processStage(stage Stage, parentAtomic *bool) error {

	descriptor(stage)

	// Если родительский этап не атомарен, проверяем флаг атомарности текущего этапа
	if parentAtomic == nil {
		if stage.Atomic == nil {
			logMessage("DEBUG", "STAGE. Atomic flag is not specified")
		} else if *stage.Atomic {
			*ATOMIC_STAGE = true
			logMessage("DEBUG", "ONLY STAGE Atomic is true")
		} else {
			*ATOMIC_STAGE = false
			logMessage("DEBUG", "STAGE Atomic is false. Parent = N/A. Skip...")
		}
	} else if *parentAtomic {
		if stage.Atomic == nil {
			*ATOMIC_STAGE = true
			logMessage("DEBUG", "STAGE. Atomic flag is not specified. Parent = TRUE")
		} else if *stage.Atomic {
			*ATOMIC_STAGE = true
			logMessage("DEBUG", "STAGE Atomic is true. Skip...")
		} else {
			logMessage("DEBUG", "ONLY STAGE Atomic is false")
		}
	} else {
		if stage.Atomic == nil {
			*ATOMIC_STAGE = false
			logMessage("DEBUG", "STAGE. Atomic flag is not specified. Parent = FASLE. Skip...")
		} else if *stage.Atomic {
			*ATOMIC_STAGE = true
			logMessage("DEBUG", "ONLY STAGE Atomic is true")
		} else {
			logMessage("DEBUG", "STAGE Atomic is false. PARENT = FALSE. Skip...")
		}
	}

	// Выполняем предварительные проверки, если они указаны
	if err := checkHandler(stage, "preCheck", ATOMIC_STAGE); err != nil {
		if *ATOMIC_STAGE {
			return err
		}
	}

	// Выполняем предварительные действия (скрипты), если они указаны
	if err := actionHandler(stage, "preScript", ATOMIC_STAGE); err != nil {
		return err
	}

	// Последовательно обрабатываем каждый шаг этапа
	for _, step := range stage.Steps {
		if err := processStage(step, parentAtomic); err != nil {
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
