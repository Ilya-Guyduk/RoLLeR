package run

import (
	"fmt"
	"io/ioutil"

	"github.com/Ilya-Guyduk/RoLLeR/handlers/plugin"
	"gopkg.in/yaml.v3"
)

const (
	MS_VERSION = "0.0.1"
)

type ActionMap struct {
	Name     string
	Action   map[string]interface{}
	Rollback map[string]interface{}
}

type MigrationSet struct {
	StandsFile          *StandsFile
	PluginController    *plugin.PluginController
	ActionMap           *map[string]ActionMap
	MigrationSetVersion string   `yaml:"msVersion"`
	Atomic              *bool    `yaml:"atomic"` // Флаг атомарности
	YAMLStandFile       string   `yaml:"stands"` // Путь к файлу стендов
	FromRelease         string   `yaml:"from_release"`
	ToRelease           string   `yaml:"to_release"`
	Stages              []Stages `yaml:"stages"` // Список этапов
}

// Метод инициализации MigrationSet
func (mg *MigrationSet) InitMigrationSet(migrationYamlFile string, pluginController *plugin.PluginController) (*MigrationSet, error) {

	fmt.Printf("[MigrationSet] migration file: %s\n", migrationYamlFile)

	// Читаем миграционный файл.
	migrationSet := &MigrationSet{}
	err := unmarshalYamlFile(migrationYamlFile, migrationSet)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal migration YAML file %s: %v", migrationSet, err)
	}
	// Читаем файл стендов из конфигурации миграции.
	stand := &StandsFile{}
	err = unmarshalYamlFile(migrationSet.YAMLStandFile, stand)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal stand YAML file %s: %v", stand, err)
	}

	// Создаем новый экземпляр MigrationSet с заполненными данными.
	newMg := &MigrationSet{
		StandsFile:       stand,
		PluginController: pluginController,
	}

	return newMg, nil
}

func (ms *MigrationSet) CheckValideData(migrationSet MigrationSet) error {

	fmt.Printf("[MigrationSet] Start valide\n")

	/*/ Валидация версий миграции и файла стендов
	if migrationSet.Migration.msVersion != MS_VERSION || migrationSet.Stands.msVersion != MS_VERSION {
		return fmt.Errorf("Unsupported version. Migration: %s, Stands: %s", migrationSet.Migration.msVersion, migrationSet.Stands.msVersion)
	} else {
		fmt.Printf("[MigrationSet] correct version\n")
	}*/

	fmt.Printf("[MigrationSet] Start valide Stands\n")
	standsErr := migrationSet.StandsFile.CheckValideData(*migrationSet.StandsFile)
	if standsErr != nil {
		return standsErr
	}

	return nil
}

func (ms *MigrationSet) UpdateRelease(migrationSet MigrationSet) error {

	for _, stage := range migrationSet.Stages {

		err := stage.ExecStage(stage, *ms.StandsFile, *ms.ActionMap, ms.Atomic, "")
		if err != nil {
			return nil
		}

	}

	return nil
}

func (mg *MigrationSet) RollbackRelease(targetRelease string) error {

	return nil
}

// UnmarshalYamlFile загружает данные из YAML файла и возвращает объект.
func unmarshalYamlFile(filePath string, target interface{}) error {
	yamlData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %v", err)
	}

	// Декодируем данные YAML в переданную структуру.
	err = yaml.Unmarshal(yamlData, target)
	if err != nil {
		return fmt.Errorf("error parsing YAML: %v", err)
	}

	return nil
}
