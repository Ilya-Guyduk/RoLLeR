package run

import (
	"fmt"
)

// Stage представляет этап обработки с его параметрами.
type Stage struct {
	Name        string      `yaml:"name"`       // Имя этапа
	Description string      `yaml:"desc"`       // Описание этапа
	Dependence  interface{} `yaml:"dependence"` // Зависимости этапа
	Atomic      *bool       `yaml:"atomic"`     // Флаг атомарности: если true, этап останавливается при ошибке
	PreCheck    Check       `yaml:"pre_check"`  // Предварительная проверка перед выполнением этапа
	PreScript   Script      `yaml:"pre_script"` // Предварительный скрипт перед выполнением этапа
	Task        Task        `yaml:"task"`
	PostCheck   Check       `yaml:"post_check"`   // Пост-проверка после выполнения этапа
	PostScript  Script      `yaml:"post_scriprt"` // Пост-скрипт после выполнения этапа
	Rollback    bool        `yaml:"rollback"`     // Флаг отката: если true, позволяет откатить изменения
	Steps       []Stage     `yaml:"stage"`        // Шаги, которые входят в этот этап
}

var ATOMIC_STAGE *bool // Глобальный флаг атомарности текущего этапа

func descriptor(stage Stage) {
	logMessage("INFO", fmt.Sprintf("========== Start stage: %s ==========", stage.Name))

	// Если указано описание этапа, выводим его в лог
	if stage.Description != "" {
		logMessage("INFO", fmt.Sprintf("========== %s", stage.Description))
	}
}

func isFlagSpecified(flag *bool) (bool, bool) {
	if flag == nil {
		return false, false // Флаг не указан
	}
	return true, *flag // Флаг указан, возвращаем его значение
}

// processStage обрабатывает этап (Stage) с различными проверками и скриптами.
func processStage(stage Stage, parentAtomic *bool, parentName string) error {

	var ATOMIC_STAGE = new(bool)
	var stageName string

	if parentName != "" {
		stageName = parentName + "." + stage.Name
	} else {
		stageName = stage.Name
	}

	descriptor(stage)

	// Проверяем флаги `stage.Atomic` и `parentAtomic`
	stageAtomicSpecified, stageAtomicValue := isFlagSpecified(stage.Atomic)
	parentAtomicSpecified, parentAtomicValue := isFlagSpecified(parentAtomic)

	// Вычисляем итоговое значение флага атомарности для текущего этапа
	switch {
	case stageAtomicSpecified && !stageAtomicValue:
		// Если указано, что текущий этап не атомарный
		logMessage("DEBUG", fmt.Sprintf("[%s] STAGE Atomic explicitly set to false", stageName))
		*ATOMIC_STAGE = false

	case !parentAtomicSpecified && !stageAtomicSpecified:
		// Ни родительский, ни текущий флаги не указаны
		logMessage("DEBUG", fmt.Sprintf("[%s] STAGE and Parent: Atomic flags are not specified", stageName))
		*ATOMIC_STAGE = false // По умолчанию не атомарный

	case !parentAtomicSpecified && stageAtomicSpecified:
		// Указан только текущий флаг
		logMessage("DEBUG", fmt.Sprintf("[%s] STAGE Atomic specified: %v, Parent: N/A", stageName, stageAtomicValue))
		*ATOMIC_STAGE = stageAtomicValue

	case parentAtomicSpecified && !stageAtomicSpecified:
		// Указан только родительский флаг
		logMessage("DEBUG", fmt.Sprintf("[%s]STAGE: Atomic flag not specified, Parent Atomic: %v", stageName, parentAtomicValue))
		*ATOMIC_STAGE = parentAtomicValue

	case parentAtomicSpecified && stageAtomicSpecified:
		// Указаны оба флага
		logMessage("DEBUG", fmt.Sprintf("[%s] STAGE Atomic: %v, Parent Atomic: %v", stageName, stageAtomicValue, parentAtomicValue))
		*ATOMIC_STAGE = parentAtomicValue && stageAtomicValue
	}

	// Выводим итоговое значение атомарности
	logMessage("INFO", fmt.Sprintf("[%s] ATOMIC_STAGE: %v", stageName, *ATOMIC_STAGE))

	// Выполняем предварительные проверки
	if err := handler(stage, "preCheck", ATOMIC_STAGE, stageName); err != nil {
		if *ATOMIC_STAGE {
			return err
		}
	}

	// Выполняем предварительные действия (скрипты)
	if err := handler(stage, "preScript", ATOMIC_STAGE, stageName); err != nil {
		if *ATOMIC_STAGE {
			return err
		}
	}

	// Обрабатываем вложенные шаги
	for _, step := range stage.Steps {
		if err := processStage(step, ATOMIC_STAGE, stageName); err != nil {
			if *ATOMIC_STAGE {
				return err
			}
		}
	}

	// Выполняем предварительные действия (скрипты)
	if err := handler(stage, "Task", ATOMIC_STAGE, stageName); err != nil {
		if *ATOMIC_STAGE {
			return err
		}
	}

	// Выполняем пост-проверки
	if err := handler(stage, "postCheck", ATOMIC_STAGE, stageName); err != nil {
		if *ATOMIC_STAGE {
			return err
		}
	}

	// Выполняем пост-действия (скрипты)
	if err := handler(stage, "postScript", ATOMIC_STAGE, stageName); err != nil {
		if *ATOMIC_STAGE {
			return err
		}
	}

	return nil
}
