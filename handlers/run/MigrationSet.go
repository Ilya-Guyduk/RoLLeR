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
	ActionMap           map[int]ActionMap
	MigrationSetVersion string   `yaml:"msVersion"`
	Atomic              *bool    `yaml:"atomic"` // Флаг атомарности
	YAMLStandFile       string   `yaml:"stands"` // Путь к файлу стендов
	FromRelease         string   `yaml:"from_release"`
	ToRelease           string   `yaml:"to_release"`
	Stages              []Stages `yaml:"stages"` // Список этапов
}

// Метод инициализации MigrationSet
func (mg *MigrationSet) InitMigrationSet(migrationYamlFile string, pc *plugin.PluginController) (*MigrationSet, error) {

	logMessage("DEBUG", fmt.Sprintf("[MigrationSet] migration file: %s", migrationYamlFile))

	if pc == nil {
		return nil, fmt.Errorf("[MigrationSet] pluginController is nil")
	}

	// Читаем миграционный файл.
	migrationSet := &MigrationSet{}
	err := unmarshalYamlFile(migrationYamlFile, migrationSet)
	if err != nil {
		return nil, fmt.Errorf("[MigrationSet] failed to unmarshal migration YAML file %s: %v", migrationSet, err)
	}
	// Читаем файл стендов из конфигурации миграции.
	stand := &StandsFile{}
	err = unmarshalYamlFile(migrationSet.YAMLStandFile, stand)
	if err != nil {
		return nil, fmt.Errorf("[MigrationSet] failed to unmarshal stand YAML file %s: %v", stand, err)
	}
	// Создаем новый экземпляр MigrationSet с заполненными данными.
	newMg := &MigrationSet{
		StandsFile:          stand,
		PluginController:    pc,
		ActionMap:           make(map[int]ActionMap),
		MigrationSetVersion: migrationSet.MigrationSetVersion,
		Atomic:              migrationSet.Atomic,
		FromRelease:         migrationSet.FromRelease,
		ToRelease:           migrationSet.ToRelease,
		Stages:              migrationSet.Stages,
	}

	return newMg, nil
}

func (ms *MigrationSet) CheckValideData(mSet MigrationSet) error {

	logMessage("DEBUG", "[MigrationSet] Check valide...")

	if mSet.StandsFile == nil {
		return fmt.Errorf("[MigrationSet] StandsFile is empty")
	}
	if mSet.PluginController == nil {
		return fmt.Errorf("[MigrationSet] PluginController is empty")
	}
	if mSet.ActionMap == nil {
		return fmt.Errorf("[MigrationSet] ActionMap is empty")
	}
	if mSet.MigrationSetVersion == "" {
		return fmt.Errorf("[MigrationSet] msVersion is empty")
	}
	if mSet.FromRelease == "" {
		return fmt.Errorf("[MigrationSet] from_release is empty")
	}
	if mSet.ToRelease == "" {
		return fmt.Errorf("[MigrationSet] to_release is empty")
	}
	if len(mSet.Stages) == 0 {
		return fmt.Errorf("[MigrationSet] stages is empty")
	}
	logMessage("DEBUG", fmt.Sprintf("[MigrationSet] StandsFile:%s, ", mSet.StandsFile))
	logMessage("DEBUG", fmt.Sprintf("[MigrationSet] PluginController => Version:%s, ", mSet.PluginController.ControllerVersion))
	logMessage("DEBUG", fmt.Sprintf("[MigrationSet] msVersion:%s, atomic:%s, from_release:%s, to_release:%s, count stages:%s",
		mSet.MigrationSetVersion,
		mSet.Atomic,
		mSet.FromRelease,
		mSet.ToRelease,
		len(mSet.Stages)))
	/*/ Валидация версий миграции и файла стендов
	if migrationSet.Migration.msVersion != MS_VERSION || migrationSet.Stands.msVersion != MS_VERSION {
		return fmt.Errorf("Unsupported version. Migration: %s, Stands: %s", migrationSet.Migration.msVersion, migrationSet.Stands.msVersion)
	} else {
		fmt.Printf("[MigrationSet] correct version\n")
	}*/

	logMessage("DEBUG", fmt.Sprintf("[MigrationSet] Starting valide StandsFile"))
	standsErr := mSet.StandsFile.CheckValideData(*mSet.StandsFile, mSet.PluginController)
	if standsErr != nil {
		return standsErr
	}

	for _, Stage := range mSet.Stages {
		logMessage("DEBUG", fmt.Sprintf("[MigrationSet] Starting valide Stage %s", Stage.Name))
		stageErr := Stage.CheckValideData(Stage, mSet.PluginController, mSet.StandsFile)
		if stageErr != nil {
			return stageErr
		}
	}

	return nil
}

func (ms *MigrationSet) UpdateRelease(migrationSet *MigrationSet) error {
	logMessage("DEBUG", fmt.Sprintf("[MigrationSet] Update Release %s => %s", migrationSet.FromRelease, migrationSet.ToRelease))

	for _, stage := range migrationSet.Stages {
		logMessage("DEBUG", fmt.Sprintf("[MigrationSet] Start ExecStage for %s", stage.Name))
		err := stage.ExecStage(stage, migrationSet, migrationSet.Atomic, "")
		if err != nil {
			return nil
		}

	}

	return nil
}

func (ms *MigrationSet) RollbackRelease(targetRelease string) error {

	return nil
}

func (ms *MigrationSet) PutAction(Name string, Action map[string]interface{}, Rollback map[string]interface{}) (int, error) {

	return 0, nil
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
