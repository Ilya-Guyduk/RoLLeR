package run

import (
	"flag"
	"fmt"

	"github.com/Ilya-Guyduk/RoLLeR/handlers/plugin"
)

const (
	MainBanner = `
  ___     _    _        ___        __   __   _ 
 | _ \___| |  | |   ___| _ \ __ __/  \ /  \ / |
 |   / _ \ |__| |__/ -_)   / \ V / () | () || |
 |_|_\___/____|____\___|_|_\  \_/ \__(_)__(_)_|    
===============================================                                                                                             
`
)

var (
	CONFIG_PATH  string
	RUN_FLAG     bool
	DRY_RUN_FLAG bool
	PLUGINS_PATH string
)

// LoggingConfig описывает параметры логирования
type LoggingConfig struct {
	Level     string `yaml:"level"`
	Formatter string `yaml:"formatter"`
}

// Plugin описывает параметры плагинов
type PluginConfig struct {
	PluginPath     string `yaml:"plugin_path"`
	PluginRepoPath string `yaml:"plugin_repo_path"`
	DefaultRepo    string `yaml:"default_repo"`
}

type Pei struct {
	Version string `yaml:"version"`
}

// Global глобальные настройки
type Global struct {
	Logging LoggingConfig `yaml:"logging"`
	Plugin  PluginConfig  `yaml:"plugin"`
	Pei     Pei           `yaml:"pei"`
}

// rollerConfig структура конфигурации
type RollerConfig struct {
	Global Global `yaml:"global"`
}

// HandleRun обрабатывает подкоманду run
func HandleRun(args []string) error {

	// Вызов логотипа
	fmt.Printf(MainBanner)

	// Инициализация флагов
	runCmd, configPath, migrationPath, pluginsPath := setupFlags()
	if err := runCmd.Parse(args); err != nil {
		return fmt.Errorf("Error parsing flags: %w", err)
	}

	// Валидируем конфигурацию
	_, loggingData, _, err := validateYAMLConfig(*configPath)
	if err != nil {
		return fmt.Errorf("Error validating YAML config file: %s", err)
	}
	//logMessage("INFO", fmt.Sprintf("Configuration: %v", configData))
	//logMessage("INFO", fmt.Sprintf("Plugin Configuration: %v", pluginConfig))
	setupLogging(loggingData)

	logMessage("INFO", "RoLLeR Starting...")

	pc := &plugin.PluginController{}
	logMessage("DEBUG", "Creating PluginController...")
	pc, pluginErr := pc.InitPluginController(*pluginsPath, "./repo", "https://github.com/Ilya-Guyduk/RoLLeRHub/raw/main/index.json")
	if pluginErr != nil {
		logMessage("ERROR", fmt.Sprintf("Error InitPluginController: %s", pluginErr))
	} else {
		logMessage("DEBUG", fmt.Sprintf("PluginController: %s", pc))
	}

	var migrationSet *MigrationSet
	// Инициализация MigrationSet
	logMessage("INFO", fmt.Sprintf("Creating MigrationSet: %s, %s", migrationPath, pc))
	migrationSet, migrationErr := migrationSet.InitMigrationSet(*migrationPath, pc)
	if migrationErr != nil {
		logMessage("ERROR", fmt.Sprintf("Error InitMigrationSet: %s", migrationErr))
	}

	// Каскадная валидация миграции
	logMessage("INFO", fmt.Sprintf(" Starting migrationSet.CheckValideData"))
	validErr := migrationSet.CheckValideData(*migrationSet)
	if validErr != nil {
		logMessage("ERROR", fmt.Sprintf("Error CheckValideData: %s", validErr))
	}

	updateErr := migrationSet.UpdateRelease(*migrationSet)
	if updateErr != nil {
		logMessage("ERROR", fmt.Sprintf("Error UpdateRelease: %s", updateErr))
	}

	/*/ Обрабатываем манифест
	if err := migrationHandler(*migrationConfig); err != nil {
		logMessage("ERROR", fmt.Sprintf("Error processing manifest: %s", err))
	}
	*/
	defer logMessage("INFO", "RoLLer runner finished")
	return nil
}

// setupFlags инициализирует флаги командной строки
func setupFlags() (*flag.FlagSet, *string, *string, *string) {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	configPath := runCmd.String("config", "./config.yml", "Path to the YAML configuration file")
	migrationPath := runCmd.String("migration", "./migration.yml", "Path to the YAML migration file")
	pluginsPath := runCmd.String("pluginsPath", "./plugins", "Path to the plugins directory")
	return runCmd, configPath, migrationPath, pluginsPath
}
