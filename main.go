package main

import (
	"flag"
	"fmt"
)

var DRY_RUN_FLAG bool
var PLUGINS_PATH string

func main() {
	runFlag := flag.Bool("run", false, "Run the configuration steps")
	dryRunFlag := flag.Bool("dryrun", false, "Dryrun the configuration steps")
	configPath := flag.String("config", "roller.yml", "Path to the YAML configuration file")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	logFlag := flag.String("loglevel", "INFO", "Set the log level (DEBUG, INFO, ERROR)")
	pluginsPath := flag.String("pluginsPath", "./plugins", "Plugins path")

	flag.Parse()

	DRY_RUN_FLAG = *dryRunFlag
	PLUGINS_PATH = *pluginsPath

	// Настройка логирования
	setupLogging(*logFlag)
	logMessage("INFO", "RoLLer STARTING ...")

	// Загрузка плагинов исполнителя
	// Каталог с плагинами по умолчанию: ./plugins
	err := loadExecutorPlugins(PLUGINS_PATH)
	if err != nil {
		logMessage("ERROR", fmt.Sprintf("Error loading executor plugins: %v", err))
	} else {
		logMessage("INFO", "Loaded Executor plugins")
	}

	// Валидация YML манифеста
	config, err := validateYAML(*configPath)
	if err != nil {
		logMessage("ERROR", fmt.Sprintf("Error validating YAML file: %v", err))
		return
	}
	if config.Version != "" {
		logMessage("INFO", "Release version: %s", config.Version)
	}

	// Запуск роллера с флагом "--run"
	if *runFlag {

		for _, stage := range config.Stages {
			if err := processStage(stage); err != nil {
				logMessage("ERROR", fmt.Sprintf("Error processing stage %s: %v", stage.Name, err))
			}
		}

		if *verbose {
			logMessage("INFO", "Configuration processed successfully.")
		}
	}

	if *dryRunFlag {

		for _, stage := range config.Stages {
			if err := processStage(stage); err != nil {
				logMessage("ERROR", fmt.Sprintf("Error processing stage %s: %v", stage.Name, err))
			}
		}
	}
}
