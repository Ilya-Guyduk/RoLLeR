package main

import (
	"flag"
	"fmt"
)

var DRY_RUN_FLAG bool

func main() {
	runFlag := flag.Bool("run", false, "Run the configuration steps")
	dryRunFlag := flag.Bool("dryrun", false, "Dryrun the configuration steps")
	configPath := flag.String("config", "roller.yml", "Path to the YAML configuration file")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	logFlag := flag.String("loglevel", "INFO", "Set the log level (DEBUG, INFO, ERROR)")

	flag.Parse()

	DRY_RUN_FLAG = *dryRunFlag

	// Настройка логирования
	setupLogging(*logFlag)

	// Запуск роллера с флагом "--run"
	if *runFlag {
		logMessage("INFO", "RoLLer STARTING ...")

		// Валидация YML манифеста
		config, err := validateYAML(*configPath)
		if err != nil {
			logMessage("ERROR", fmt.Sprintf("Error validating YAML file: %v", err))
			return
		}

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
		logMessage("INFO", "RoLLer STARTING ///")

		config, err := validateYAML(*configPath)
		if err != nil {
			logMessage("ERROR", fmt.Sprintf("Error validating YAML file: %v", err))
		}

		for _, stage := range config.Stages {
			if err := processStage(stage); err != nil {
				logMessage("ERROR", fmt.Sprintf("Error processing stage %s: %v", stage.Name, err))
			}
		}
	}
}
