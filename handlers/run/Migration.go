package run

import (
	"errors"
	"fmt"

	"github.com/Ilya-Guyduk/RoLLeR/handlers/plugin"
)

type ActionMap struct {
	Name     string
	Action   map[string]interface{}
	Rollback map[string]interface{}
}

// Основная структура для манифеста
type Migration struct {
	Set           *MigrationSet
	ActionMap     *map[string]ActionMap
	msVersion     string   `yaml:"msVersion"`
	Atomic        *bool    `yaml:"atomic"` // Флаг атомарности
	YAMLStandFile string   `yaml:"stands"` // Путь к файлу стендов
	FromRelease   string   `yaml:"from_release"`
	ToRelease     string   `yaml:"to_release"`
	Stages        []Stages `yaml:"stages"` // Список этапов
}

func (m *Migration) CheckValideData(migration Migration) error {
	if *m.Atomic {

	}
	// Проверка уникальности имен компонентов
	nameSet := make(map[string]bool)
	for _, Stage := range migration.Stages {
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
	return nil
}

func (m *Migration) ExecMigration(migration Migration, stands StandsFile, pc *plugin.PluginController) error {

	for _, stage := range migration.Stages {

		err := stage.ExecStage(stage, stands, *m.ActionMap, stage.Atomic, "")
		if err != nil {
			return nil
		}

	}

	return nil
}

func (m *Migration) RollbackMigration() error {

	return nil
}

func (m *Migration) checkAtomic() error {
	// Логируем значение флага атомарности
	if m.Atomic == nil {
		logMessage("INFO", "Migration Atomic flag is not specified")
		return nil
	} else if *m.Atomic {
		logMessage("INFO", "Migration Atomic is true")
		return nil
	} else {
		logMessage("INFO", "Migration Atomic is false")
		return nil
	}
}

// Глобальные переменные
var ATOMIC_MIGRATION string
var YAML_STAND_FILE string

// Основная функция для обработки манифеста
func migrationHandler(migration Migration) error {

	// Логируем версию релиза, если указана
	if migration.msVersion != "" {
		logMessage("INFO", fmt.Sprintf("Migration version: %s", migration.msVersion))
	}

	// Логируем версию релиза, если указана
	if migration.FromRelease != "" && migration.ToRelease != "" {
		logMessage("INFO", fmt.Sprintf("Migration: %s => %s", migration.FromRelease, migration.ToRelease))
	} else {
		return errors.New(fmt.Sprintf("Migration endpoints is empty: from_release: %s, to_release: %s", migration.FromRelease, migration.ToRelease))
	}

	logFunc := migration.checkAtomic()
	if logFunc != nil {
		fmt.Println("Error:", logFunc)
	}

	// Сохраняем путь к файлу стендов
	if migration.YAMLStandFile != "" {
		YAML_STAND_FILE = migration.YAMLStandFile
		logMessage("INFO", fmt.Sprintf("YAMLStandFile: %s", migration.YAMLStandFile))
	}

	// Обработка этапов манифеста
	logMessage("INFO", "Processing stages in Migration")
	for _, stage := range migration.Stages {
		if err := processStage(stage, migration.Atomic, ""); err != nil {
			logMessage("ERROR", fmt.Sprintf("Error processing stage %s: %v", stage.Name, err))
			if migration.Atomic != nil && *migration.Atomic {
				return err
			}
		}
	}

	return nil
}
