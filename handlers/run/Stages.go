package run

import (
	"fmt"

	"github.com/Ilya-Guyduk/RoLLeR/handlers/plugin"
)

var ATOMIC_STAGE *bool // Глобальный флаг атомарности текущего этапа

// Stage представляет этап обработки с его параметрами.
type Stages struct {
	//Set         *MigrationSet
	Name        string      `yaml:"name"`       // Имя этапа
	Description string      `yaml:"desc"`       // Описание этапа
	Dependence  interface{} `yaml:"dependence"` // Зависимости этапа
	Atomic      *bool       `yaml:"atomic"`     // Флаг атомарности: если true, этап останавливается при ошибке
	PreCheck    []Check     `yaml:"pre_check"`  // Предварительная проверка перед выполнением этапа
	PreScript   []Script    `yaml:"pre_script"` // Предварительный скрипт перед выполнением этапа
	Task        []Task      `yaml:"task"`
	PostCheck   []Check     `yaml:"post_check"`   // Пост-проверка после выполнения этапа
	PostScript  []Script    `yaml:"post_scriprt"` // Пост-скрипт после выполнения этапа
	Rollback    bool        `yaml:"rollback"`     // Флаг отката: если true, позволяет откатить изменения
	Stages      []Stages    `yaml:"stages"`       // Шаги, которые входят в этот этап
}

func (s *Stages) CheckValideData(stage Stages, pc *plugin.PluginController, stands StandsFile, logMessage func(string, string, ...interface{})) error {

	logMessage("DEBUG", fmt.Sprintf("[Stage:'%s']>[Valid] Start validation", stage.Name))

	//if len(stage.Task) == 0 && len(stage.Stages) == 0 {
	//	return fmt.Errorf("[Stages > %s]>[Valid] 'task' and 'stages' is empty", stage.Name)
	//}

	//Валидация вложенных этапов, если они есть
	if len(stage.PreCheck) != 0 {
		// Проверка уникальности имен компонентов
		nameSet := make(map[string]bool)
		for _, PreCheck := range stage.PreCheck {
			// Проверяем, что имя компонента не пустое
			if PreCheck.Name == "" {
				return fmt.Errorf("[Stage:'%s']>[Valid] PreCheck.Name is empty", stage.Name)
			}

			// Проверяем уникальность имени компонента
			if nameSet[PreCheck.Name] {
				return fmt.Errorf("[Stage:'%s']>[Valid] duplicate Stage.Name found: %s", stage.Name, PreCheck.Name)
			}
			nameSet[PreCheck.Name] = true

			logMessage("DEBUG", fmt.Sprintf("[Stage:'%s']>[Valid] Starting CheckValideData PreCheck: '%s'", stage.Name, PreCheck.Name))
			// Проверяем остальные данные компонента
			_, _, PreCheckErr := PreCheck.CascadeValidation(PreCheck, pc, stands, logMessage)
			if PreCheckErr != nil {
				return PreCheckErr
			}
			logMessage("DEBUG", fmt.Sprintf("[Stage:'%s']>[Valid] CheckValideData PreCheck '%s' finish!", stage.Name, PreCheck.Name))
		}
	} else {
		logMessage("DEBUG", fmt.Sprintf("[Stage:'%s']>[Valid] Missing PreCheck...", stage.Name))
	}

	//Валидация вложенных этапов, если они есть
	if stage.PreScript != nil {
		logMessage("DEBUG", fmt.Sprintf("[Stages > %s]>[Valid] Check PreScript.", stage.Name))
		// Проверка уникальности имен компонентов
		nameSet := make(map[string]bool)
		for _, PreScript := range stage.PreScript {
			// Проверяем, что имя компонента не пустое
			if PreScript.Name == "" {
				return fmt.Errorf("PreCheck.Name is empty")
			}

			// Проверяем уникальность имени компонента
			if nameSet[PreScript.Name] {
				return fmt.Errorf("duplicate Stage.Name found: %s", PreScript.Name)
			}
			nameSet[PreScript.Name] = true

			logMessage("DEBUG", fmt.Sprintf("[Stages > %s]>[Valid] Starting CheckValideData PreScript: '%s'", stage.Name, PreScript.Name))
			// Проверяем остальные данные компонента
			_, _, PreScriptErr := PreScript.CascadeValidation(PreScript, pc, stands, logMessage)
			if PreScriptErr != nil {
				return PreScriptErr
			}
		}
	} else {
		logMessage("DEBUG", fmt.Sprintf("[Stages > %s]>[Valid] Missing PreScript...", stage.Name))
	}

	//Валидация вложенных этапов, если они есть
	if len(stage.Stages) != 0 {
		// Проверка уникальности имен компонентов
		nameSet := make(map[string]bool)
		for _, Stage := range stage.Stages {
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
			componentErr := Stage.CheckValideData(Stage, pc, stands, logMessage)
			if componentErr != nil {
				return componentErr
			}
		}
	} else {
		logMessage("DEBUG", fmt.Sprintf("[Stages > %s]>[Valid] Missing Stages...", stage.Name))
	}

	//Валидация вложенных этапов, если они есть
	if len(stage.PostCheck) != 0 {
		// Проверка уникальности имен компонентов
		nameSet := make(map[string]bool)
		for _, PostCheck := range stage.PostCheck {
			// Проверяем, что имя компонента не пустое
			if PostCheck.Name == "" {
				return fmt.Errorf("component.Name is empty")
			}

			// Проверяем уникальность имени компонента
			if nameSet[PostCheck.Name] {
				return fmt.Errorf("duplicate Stage.Name found: %s", PostCheck.Name)
			}
			nameSet[PostCheck.Name] = true

			// Проверяем остальные данные компонента
			_, _, componentErr := PostCheck.CascadeValidation(PostCheck, pc, stands, logMessage)
			if componentErr != nil {
				return componentErr
			}
		}
	} else {
		logMessage("DEBUG", fmt.Sprintf("[Stages > %s]>[Valid] Missing PostCheck...", stage.Name))
	}

	//Валидация вложенных этапов, если они есть
	if len(stage.PostScript) != 0 {
		// Проверка уникальности имен компонентов
		nameSet := make(map[string]bool)
		for _, PostScript := range stage.PostScript {
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
			_, _, PreScriptErr := PostScript.CascadeValidation(PostScript, pc, stands, logMessage)
			if PreScriptErr != nil {
				return PreScriptErr
			}
		}
	} else {
		logMessage("DEBUG", fmt.Sprintf("[Stages > %s]>[Valid] Missing PostScript...", stage.Name))
	}
	return nil
}

func (s *Stages) CheckMyAtomic(stageName string, myAtomic *bool, parentAtomic *bool, logMessage func(string, string, ...interface{})) *bool {
	var atomFlag *bool = new(bool)

	// Проверяем флаги `stage.Atomic` и `parentAtomic`
	stageAtomicSpecified, stageAtomicValue := isFlagSpecified(myAtomic)
	parentAtomicSpecified, parentAtomicValue := isFlagSpecified(parentAtomic)

	// Вычисляем итоговое значение флага атомарности для текущего этапа
	switch {
	case stageAtomicSpecified && !stageAtomicValue:
		// Если указано, что текущий этап не атомарный
		logMessage("DEBUG", fmt.Sprintf("[Stage > %s] STAGE Atomic explicitly set to false", stageName))
		*atomFlag = false

	case !parentAtomicSpecified && !stageAtomicSpecified:
		// Ни родительский, ни текущий флаги не указаны
		logMessage("DEBUG", fmt.Sprintf("[Stage > %s] STAGE and Parent: Atomic flags are not specified", stageName))
		*atomFlag = false // По умолчанию не атомарный

	case !parentAtomicSpecified && stageAtomicSpecified:
		// Указан только текущий флаг
		logMessage("DEBUG", fmt.Sprintf("[Stage > %s] STAGE Atomic specified: %v, Parent: N/A", stageName, stageAtomicValue))
		*atomFlag = stageAtomicValue

	case parentAtomicSpecified && !stageAtomicSpecified:
		// Указан только родительский флаг
		logMessage("DEBUG", fmt.Sprintf("[Stage > %s]STAGE: Atomic flag not specified, Parent Atomic: %v", stageName, parentAtomicValue))
		*atomFlag = parentAtomicValue

	case parentAtomicSpecified && stageAtomicSpecified:
		// Указаны оба флага
		logMessage("DEBUG", fmt.Sprintf("[Stage > %s] STAGE Atomic: %v, Parent Atomic: %v", stageName, stageAtomicValue, parentAtomicValue))
		*atomFlag = parentAtomicValue && stageAtomicValue
	}

	// Выводим итоговое значение атомарности
	logMessage("INFO", fmt.Sprintf("[Stage > %s] ATOMIC_STAGE: %v", stageName, atomFlag))
	return atomFlag
}

func (s *Stages) ExecStage(stage Stages, ms *MigrationSet, parentAtomic *bool, parentName string, logMessage func(string, string, ...interface{})) error {
	// Создаём локальную переменную для хранения атомарности текущего этапа
	//var ATOMIC_STAGE = new(bool)
	stageName := s.setName(parentName, stage.Name)

	logMessage("INFO", fmt.Sprintf("========== Start stage: %s ==========", stage.Name))

	// Если указано описание этапа, выводим его в лог
	if stage.Description != "" {
		logMessage("INFO", fmt.Sprintf("========== %s", stage.Description))
	}

	// Проверяем и вычисляем атомарность этапа
	MY_ATOMIC_STAGE := stage.CheckMyAtomic(stageName, stage.Atomic, parentAtomic, logMessage)

	// Шаг 1: Выполняем PreCheck, если он указан
	if stage.PreCheck != nil {

		for _, PreCheck := range stage.PreCheck {

			if err := PreCheck.ExecCheck(PreCheck, stageName, ms.PluginController, ms.StandsFile, logMessage); err != nil {

				logMessage("ERROR", fmt.Sprintf("[Stage > %s] PreCheck failed: %v", stageName, err))
				return err

			} else {

				//err := ms.AddActionToGraph(PreCheck.Name, ms.Action, []PreCheck.Actions)
				//if err != nil {
				//	return nil
				//}
			}
		}
	}

	// Шаг 2: Выполняем PreScript, если он указан
	if stage.PreScript != nil {

		for _, PreScript := range stage.PreScript {
			if err := PreScript.ExecScript(PreScript, stageName, ms.PluginController, ms.StandsFile, logMessage); err != nil {
				logMessage("ERROR", fmt.Sprintf("[Stage > %s] PreScript failed: %v", stageName, err))
				if *MY_ATOMIC_STAGE {
					return err
				}
			}
		}
	}

	// Шаг 3: Выполняем вложенные этапы, если они есть
	for _, subStage := range stage.Stages {
		logMessage("INFO", fmt.Sprintf("[Stage > %s] Processing sub-stage: %s", stageName, subStage.Name))
		if err := s.ExecStage(subStage, ms, MY_ATOMIC_STAGE, stageName, logMessage); err != nil {
			logMessage("ERROR", fmt.Sprintf("[Stage > %s] Sub-stage %s failed: %v", stageName, subStage.Name, err))
			if *MY_ATOMIC_STAGE {
				return err
			}
		}
	}

	// Шаг 4: Выполняем Task, если он указан
	for _, task := range stage.Task {
		logMessage("INFO", fmt.Sprintf("[Stage > %s] Executing Task...", stageName))
		if err := task.ExecTask(task, stageName, ms.PluginController, ms.StandsFile, logMessage); err != nil {
			logMessage("ERROR", fmt.Sprintf("[Stage > %s] Task failed: %v", stageName, err))
			if *MY_ATOMIC_STAGE {
				return err
			}
		}
	}

	// Шаг 5: Выполняем PostScript, если он указан
	if stage.PostScript != nil {
		logMessage("INFO", fmt.Sprintf("[%s] Executing PostScript...", stageName))
		for _, PostScript := range stage.PostScript {
			if err := PostScript.ExecScript(PostScript, stageName, ms.PluginController, ms.StandsFile, logMessage); err != nil {
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
		for _, PostCheck := range stage.PostCheck {
			if err := PostCheck.ExecCheck(PostCheck, stageName, ms.PluginController, ms.StandsFile, logMessage); err != nil {
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

func (s *Stages) setName(parentName string, currentName string) string {

	var stageName string

	if parentName != "" {
		stageName = parentName + "." + currentName
	} else {
		stageName = currentName
	}

	return stageName
}

func isFlagSpecified(flag *bool) (bool, bool) {
	if flag == nil {
		return false, false // Флаг не указан
	}
	return true, *flag // Флаг указан, возвращаем его значение
}
