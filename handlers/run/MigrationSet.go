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

type MigrationSet struct {
	MigrationSetVersion   string
	CurrentVersionRelease string
	TargetVersionRelease  string
	Migration             *Migration
	Stands                *StandsFile
	PluginController      *plugin.PluginController
}

// Метод инициализации MigrationSet
func (mg *MigrationSet) InitMigrationSet(migrationYamlFile string, pluginController *plugin.PluginController) (*MigrationSet, error) {

	fmt.Printf("[MigrationSet] migration file: %s\n", migrationYamlFile)

	// Читаем миграционный файл.
	migration := &Migration{}
	err := unmarshalYamlFile(migrationYamlFile, migration)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal migration YAML file %s: %v", migration, err)
	}
	//if migration.msVersion != MS_VERSION {
	//	return nil, fmt.Errorf("failed to unmarshal migration YAML file:")
	//}

	// Читаем файл стендов из конфигурации миграции.
	stand := &StandsFile{}
	err = unmarshalYamlFile(migration.YAMLStandFile, stand)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal stand YAML file %s: %v", stand, err)
	}

	// Создаем новый экземпляр MigrationSet с заполненными данными.
	newMg := &MigrationSet{
		MigrationSetVersion:   "0.0.1",
		CurrentVersionRelease: migration.FromRelease,
		TargetVersionRelease:  migration.ToRelease,
		Migration:             migration,
		Stands:                stand,
		PluginController:      pluginController,
	}

	return newMg, nil
}

func (mg *MigrationSet) CheckValideData(migrationSet MigrationSet) error {

	fmt.Printf("[MigrationSet] Start valide\n")

	/*/ Валидация версий миграции и файла стендов
	if migrationSet.Migration.msVersion != MS_VERSION || migrationSet.Stands.msVersion != MS_VERSION {
		return fmt.Errorf("Unsupported version. Migration: %s, Stands: %s", migrationSet.Migration.msVersion, migrationSet.Stands.msVersion)
	} else {
		fmt.Printf("[MigrationSet] correct version\n")
	}*/

	if migrationSet.TargetVersionRelease == "" {
		return fmt.Errorf("TargetVersionRelease is empty")
	}

	fmt.Printf("[MigrationSet] Start valide Stands\n")
	standsErr := migrationSet.Stands.CheckValideData(*migrationSet.Stands)
	if standsErr != nil {
		return standsErr
	}

	fmt.Printf("[MigrationSet] Start valide Migration\n")
	migrationErr := migrationSet.Migration.CheckValideData(*migrationSet.Migration)
	if migrationErr != nil {
		return migrationErr
	}

	return nil
}

func (mg *MigrationSet) UpdateRelease(migrationSet MigrationSet) error {

	migrationErr := migrationSet.Migration.ExecMigration(*migrationSet.Migration, *migrationSet.Stands, migrationSet.PluginController)
	if migrationErr != nil {
		return migrationErr
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
