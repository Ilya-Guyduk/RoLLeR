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

var (
	DEFAULT_MS_MIN_EXEC_THREADS int = 1
	DEFAULT_MS_MAX_EXEC_THREADS int = 100
	DEFAULT_MS_EXEC_THREADS     int = 1
)

var (
	STAGE_ORDER = "consistently"
)

var (
	DEFAULT_SUPPORT_ROLLBACK = false
)

// Граф зависимостей
type DependencyGraph struct {
	Actions      map[string]Action   // Карта действий
	Dependencies map[string][]string // Зависимости для каждого действия
}

type Action struct {
	Name     string
	Action   map[string]interface{}
	Rollback map[string]interface{}
}

type MigrationSet struct {
	StandsFile          *StandsFile
	PluginController    *plugin.PluginController
	DependencyGraph     *DependencyGraph
	MigrationSetVersion string   `yaml:"msVersion"`
	Atomic              *bool    `yaml:"atomic"` // Флаг атомарности
	YAMLStandFile       string   `yaml:"stands"` // Путь к файлу стендов
	FromRelease         string   `yaml:"from_release"`
	ToRelease           string   `yaml:"to_release"`
	Stages              []Stages `yaml:"stages"` // Список этапов
}

// Метод инициализации MigrationSet
func (ms *MigrationSet) NewMigrationSet(MigrationSetYamlFile string, pc *plugin.PluginController, logMessage func(string, string, ...interface{})) (*MigrationSet, error) {

	logMessage("DEBUG", fmt.Sprintf("[MigrationSet]>[New] migration file: %s", MigrationSetYamlFile))

	if pc == nil {
		return nil, fmt.Errorf("[MigrationSet]>[New] pluginController is nil")
	}

	// Читаем миграционный файл.
	migrationSet := &MigrationSet{}
	err := unmarshalYamlFile(MigrationSetYamlFile, migrationSet)
	if err != nil {
		return nil, fmt.Errorf("[MigrationSet]>[New] Unmarshal 'migration' YAML: %v: %v", migrationSet, err)
	}
	// Читаем файл стендов из конфигурации миграции.
	stand := &StandsFile{}
	err = unmarshalYamlFile(migrationSet.YAMLStandFile, stand)
	if err != nil {
		return nil, fmt.Errorf("[MigrationSet]>[New] Unmarshal 'stands' YAML: %v", err)
	}
	// Создаем новый экземпляр MigrationSet с заполненными данными.
	newMg := &MigrationSet{
		StandsFile:          stand,
		PluginController:    pc,
		DependencyGraph:     &DependencyGraph{Actions: make(map[string]Action), Dependencies: make(map[string][]string)},
		MigrationSetVersion: migrationSet.MigrationSetVersion,
		Atomic:              migrationSet.Atomic,
		FromRelease:         migrationSet.FromRelease,
		ToRelease:           migrationSet.ToRelease,
		Stages:              migrationSet.Stages,
	}

	return newMg, nil
}

func (ms *MigrationSet) CascadeValidation(mSet MigrationSet, logMessage func(string, string, ...interface{})) error {
	err := ms.ValidateMS(mSet)
	if err != nil {
		return err
	}

	// Channel to receive errors from goroutines
	errChan := make(chan error, len(mSet.Stages)+1) // Buffered channel to avoid deadlocks

	// Start goroutine for StandsFile validation
	go func() {
		logMessage("INFO", "[MigrationSet]>[Valid] Start validation 'StandsFile'")
		errChan <- mSet.StandsFile.CascadeValidation(*mSet.StandsFile, mSet.PluginController, logMessage)
	}()

	// Start goroutines for Stage validations
	for _, Stage := range mSet.Stages {
		go func(stage Stages) {
			logMessage("INFO", fmt.Sprintf("[MigrationSet]>[Valid] Start validation 'Stage' '%s'", stage.Name))
			errChan <- stage.CheckValideData(stage, mSet.PluginController, *mSet.StandsFile, logMessage)
		}(Stage)
	}

	// Collect errors from goroutines
	for i := 0; i < len(mSet.Stages)+1; i++ {
		if err := <-errChan; err != nil {
			return err // Return the first error encountered
		}
	}

	close(errChan) // Close the channel when done
	return nil
}

func (ms *MigrationSet) ValidateMS(mSet MigrationSet) error {

	if mSet.StandsFile == nil {
		return fmt.Errorf("[MigrationSet]>[Valid] 'StandsFile' is empty")
	}
	if mSet.PluginController == nil {
		return fmt.Errorf("[MigrationSet]>[Valid] 'PluginController' is empty")
	}
	if mSet.MigrationSetVersion == "" {
		return fmt.Errorf("[MigrationSet]>[Valid] 'msVersion' is empty")
	}
	if mSet.FromRelease == "" {
		return fmt.Errorf("[MigrationSet]>[Valid] 'from_release' is empty")
	}
	if mSet.ToRelease == "" {
		return fmt.Errorf("[MigrationSet]>[Valid] 'to_release' is empty")
	}
	if len(mSet.Stages) == 0 {
		return fmt.Errorf("[MigrationSet]>[Valid] 'stages' is empty")
	}

	return nil
}

func (ms *MigrationSet) UpdateRelease(mSet *MigrationSet, logMessage func(string, string, ...interface{})) error {

	logMessage("DEBUG", fmt.Sprintf("[MigrationSet]>[Update] Update Release '%s'=>'%s'", mSet.FromRelease, mSet.ToRelease))

	for _, stage := range mSet.Stages {
		err := stage.ExecStage(stage, mSet, mSet.Atomic, "", logMessage)
		if err != nil {
			return nil
		} else {
			return err
		}

	}

	return nil
}

func (ms *MigrationSet) RollbackRelease(targetRelease string, logMessage func(string, string, ...interface{})) error {

	return nil
}

// Метод для добавления действия в граф
func (ms *MigrationSet) AddActionToGraph(actionName string, action Action, dependencies []string) error {
	if _, exists := ms.DependencyGraph.Actions[actionName]; exists {
		return fmt.Errorf("[MigrationSet]>[AddActionToGraph] Action %s already exists", actionName)
	}

	ms.DependencyGraph.Actions[actionName] = action
	ms.DependencyGraph.Dependencies[actionName] = dependencies
	return nil
}

func (ms *MigrationSet) CreateMSFiles(mSet *MigrationSet, logMessage func(string, string, ...interface{})) error {

	logMessage("DEBUG", fmt.Sprintf("[MigrationSet]>[Update] Update Release '%s'=>'%s'", mSet.FromRelease, mSet.ToRelease))

	for _, stage := range mSet.Stages {
		logMessage("DEBUG", fmt.Sprintf("[MigrationSet]>[Update] Start ExecStage for %s", stage.Name))
		err := stage.ExecStage(stage, mSet, mSet.Atomic, "", logMessage)
		if err != nil {
			return nil
		} else {
			return err
		}

	}

	return nil
}

func (ms *MigrationSet) SetPluginController(pc *plugin.PluginController) error {
	ms.PluginController = pc
	return nil
}

func (ms *MigrationSet) SetMinExecThreads(num int) {
	DEFAULT_MS_MIN_EXEC_THREADS = num
}

func (ms *MigrationSet) SetMaxExecThreads(num int) {
	DEFAULT_MS_MIN_EXEC_THREADS = num
}

func (ms *MigrationSet) GetStages() []Stages {
	return ms.Stages
}

// UnmarshalYamlFile загружает данные из YAML файла и возвращает объект.
func unmarshalYamlFile(filePath string, target interface{}) error {
	yamlData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ErrorReadYAML '%s': %v", filePath, err)
	}

	// Декодируем данные YAML в переданную структуру.
	err = yaml.Unmarshal(yamlData, target)
	if err != nil {
		return fmt.Errorf("ErrorParsYAML: %v", err)
	}

	return nil
}
