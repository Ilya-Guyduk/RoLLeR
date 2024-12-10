package run

import (
	"fmt"
)

var ATOMIC_STAGE *bool // Глобальный флаг атомарности текущего этапа
// Stage представляет этап обработки с его параметрами.
type Stages struct {
	Set         *MigrationSet
	Name        string      `yaml:"name"`       // Имя этапа
	Description string      `yaml:"desc"`       // Описание этапа
	Dependence  interface{} `yaml:"dependence"` // Зависимости этапа
	Atomic      *bool       `yaml:"atomic"`     // Флаг атомарности: если true, этап останавливается при ошибке
	PreCheck    *[]Check    `yaml:"pre_check"`  // Предварительная проверка перед выполнением этапа
	PreScript   *[]Script   `yaml:"pre_script"` // Предварительный скрипт перед выполнением этапа
	Task        *[]Task     `yaml:"task"`
	PostCheck   *[]Check    `yaml:"post_check"`   // Пост-проверка после выполнения этапа
	PostScript  *[]Script   `yaml:"post_scriprt"` // Пост-скрипт после выполнения этапа
	Rollback    bool        `yaml:"rollback"`     // Флаг отката: если true, позволяет откатить изменения
	Stages      *[]Stages   `yaml:"stage"`        // Шаги, которые входят в этот этап
}

func (s *Stages) CheckValideData(stage Stages) error {

	//Валидация вложенных этапов, если они есть
	if len(*stage.PreCheck) != 0 {
		// Проверка уникальности имен компонентов
		nameSet := make(map[string]bool)
		for _, PreCheck := range *stage.PreCheck {
			// Проверяем, что имя компонента не пустое
			if PreCheck.Name == "" {
				return fmt.Errorf("PreCheck.Name is empty")
			}

			// Проверяем уникальность имени компонента
			if nameSet[PreCheck.Name] {
				return fmt.Errorf("duplicate Stage.Name found: %s", PreCheck.Name)
			}
			nameSet[PreCheck.Name] = true

			// Проверяем остальные данные компонента
			_, _, PreCheckErr := PreCheck.CheckValideData(PreCheck)
			if PreCheckErr != nil {
				return PreCheckErr
			}
		}
	}

	//Валидация вложенных этапов, если они есть
	if len(*stage.PreScript) != 0 {
		// Проверка уникальности имен компонентов
		nameSet := make(map[string]bool)
		for _, PreScript := range *stage.PreScript {
			// Проверяем, что имя компонента не пустое
			if PreScript.Name == "" {
				return fmt.Errorf("PreCheck.Name is empty")
			}

			// Проверяем уникальность имени компонента
			if nameSet[PreScript.Name] {
				return fmt.Errorf("duplicate Stage.Name found: %s", PreScript.Name)
			}
			nameSet[PreScript.Name] = true

			// Проверяем остальные данные компонента
			PreScriptErr := PreScript.CheckValideData(PreScript)
			if PreScriptErr != nil {
				return PreScriptErr
			}
		}
	}

	//Валидация вложенных этапов, если они есть
	if len(*stage.Stages) != 0 {
		// Проверка уникальности имен компонентов
		nameSet := make(map[string]bool)
		for _, Stage := range *stage.Stages {
			// Проверяем, что имя компонента не пустое
			if Stage.Name == "" {
				return fmt.Errorf("component.Name is empty")
			}

			// Проверяем уникальность имени компонента
			if nameSet[Stage.Name] {
				return fmt.Errorf("duplicate Stage.Name found: %s", Stage.Name)
			}
			nameSet[Stage.Name] = true

			// Проверяем остальные данные компонента
			componentErr := Stage.CheckValideData(Stage)
			if componentErr != nil {
				return componentErr
			}
		}
	}

	//Валидация вложенных этапов, если они есть
	if len(*stage.PreCheck) != 0 {
		// Проверка уникальности имен компонентов
		nameSet := make(map[string]bool)
		for _, PreCheck := range *stage.PreCheck {
			// Проверяем, что имя компонента не пустое
			if PreCheck.Name == "" {
				return fmt.Errorf("component.Name is empty")
			}

			// Проверяем уникальность имени компонента
			if nameSet[PreCheck.Name] {
				return fmt.Errorf("duplicate Stage.Name found: %s", PreCheck.Name)
			}
			nameSet[PreCheck.Name] = true

			// Проверяем остальные данные компонента
			_, _, componentErr := PreCheck.CheckValideData(PreCheck)
			if componentErr != nil {
				return componentErr
			}
		}
	}

	//Валидация вложенных этапов, если они есть
	if len(*stage.PostScript) != 0 {
		// Проверка уникальности имен компонентов
		nameSet := make(map[string]bool)
		for _, PostScript := range *stage.PostScript {
			// Проверяем, что имя компонента не пустое
			if PostScript.Name == "" {
				return fmt.Errorf("PreCheck.Name is empty")
			}

			// Проверяем уникальность имени компонента
			if nameSet[PostScript.Name] {
				return fmt.Errorf("duplicate Stage.Name found: %s", PostScript.Name)
			}
			nameSet[PostScript.Name] = true

			// Проверяем остальные данные компонента
			PreScriptErr := PostScript.CheckValideData(PostScript)
			if PreScriptErr != nil {
				return PreScriptErr
			}
		}
	}
	return nil
}

func (s *Stages) CheckMyAtomic(stageName string, myAtomic *bool, parentAtomic *bool) *bool {

	// Проверяем флаги `stage.Atomic` и `parentAtomic`
	stageAtomicSpecified, stageAtomicValue := isFlagSpecified(myAtomic)
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
	return ATOMIC_STAGE
}

func (s *Stages) setName(parentName string, currentName string) string {

	var stageName string

	if parentName != "" {
		stageName = parentName + "." + currentName
	} else {
		stageName = currentName
	}

	return stageName
}

func (s *Stages) ExecStage(stage Stages, stands StandsFile, ActionMap map[string]ActionMap, parentAtomic *bool, parentName string) error {
	// Создаём локальную переменную для хранения атомарности текущего этапа
	//var ATOMIC_STAGE = new(bool)
	stageName := s.setName(parentName, stage.Name)
	descriptor(stage)

	// Проверяем и вычисляем атомарность этапа
	MY_ATOMIC_STAGE := stage.CheckMyAtomic(stageName, stage.Atomic, parentAtomic)

	// Шаг 1: Выполняем PreCheck, если он указан
	if stage.PreCheck != nil {
		logMessage("INFO", fmt.Sprintf("[%s] Executing PreCheck...", stageName))
		for _, PreCheck := range *stage.PreCheck {
			if err := PreCheck.ExecCheck(PreCheck, stageName); err != nil {
				logMessage("ERROR", fmt.Sprintf("[%s] PreCheck failed: %v", stageName, err))
				if *MY_ATOMIC_STAGE {
					return err
				}
			}
		}
	}

	// Шаг 2: Выполняем PreScript, если он указан
	if stage.PreScript != nil {
		logMessage("INFO", fmt.Sprintf("[%s] Executing PreScript...", stageName))
		for _, PreScript := range *stage.PreScript {
			if err := PreScript.ExecScript(PreScript, stageName); err != nil {
				logMessage("ERROR", fmt.Sprintf("[%s] PreScript failed: %v", stageName, err))
				if *MY_ATOMIC_STAGE {
					return err
				}
			}
		}
	}

	// Шаг 3: Выполняем вложенные этапы, если они есть
	for _, subStage := range *stage.Stages {
		logMessage("INFO", fmt.Sprintf("[%s] Processing sub-stage: %s", stageName, subStage.Name))
		if err := s.ExecStage(subStage, stands, ActionMap, MY_ATOMIC_STAGE, stageName); err != nil {
			logMessage("ERROR", fmt.Sprintf("[%s] Sub-stage %s failed: %v", stageName, subStage.Name, err))
			if *MY_ATOMIC_STAGE {
				return err
			}
		}
	}

	// Шаг 4: Выполняем Task, если он указан
	for _, task := range *stage.Task {
		logMessage("INFO", fmt.Sprintf("[%s] Executing Task...", stageName))
		if err := task.ExecTask(task, stageName); err != nil {
			logMessage("ERROR", fmt.Sprintf("[%s] Task failed: %v", stageName, err))
			if *MY_ATOMIC_STAGE {
				return err
			}
		}
	}

	// Шаг 5: Выполняем PostScript, если он указан
	if stage.PostScript != nil {
		logMessage("INFO", fmt.Sprintf("[%s] Executing PostScript...", stageName))
		for _, PostScript := range *stage.PostScript {
			if err := PostScript.ExecScript(PostScript, stageName); err != nil {
				logMessage("ERROR", fmt.Sprintf("[%s] PostScript failed: %v", stageName, err))
				if *MY_ATOMIC_STAGE {
					return err
				}
			}
		}
	}

	// Шаг 6: Выполняем PostCheck, если он указан
	if stage.PostCheck != nil {
		logMessage("INFO", fmt.Sprintf("[%s] Executing PostCheck...", stageName))
		for _, PostCheck := range *stage.PostCheck {
			if err := PostCheck.ExecCheck(PostCheck, stageName); err != nil {
				logMessage("ERROR", fmt.Sprintf("[%s] PostCheck failed: %v", stageName, err))
				if *MY_ATOMIC_STAGE {
					return err
				}
			}
		}
	}

	logMessage("INFO", fmt.Sprintf("[%s] Stage completed successfully.", stageName))
	return nil
}

func descriptor(stage Stages) {
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
