package run

import (
	"flag"
	"fmt"
)

var RUN_FLAG bool
var DRY_RUN_FLAG bool
var PLUGINS_PATH string

// HandleRun обрабатывает подкоманду run
func HandleRun(args []string) error {

	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	//dryRunFlag := runCmd.Bool("dryrun", false, "Dryrun the configuration steps")
	configPath := runCmd.String("config", "roller.yml", "Path to the YAML configuration file")
	logFlag := runCmd.String("loglevel", "INFO", "Set the log level (DEBUG, INFO, ERROR)")
	pluginsPath := runCmd.String("pluginsPath", "./plugins", "Plugins path")

	runCmd.Parse(args)

	// Настраиваем логи
	setupLogging(*logFlag)

	logMessage("INFO", "RoLLer runner starting ...")

	// Загрузка плагинов
	if err := loadExecutorPlugins(*pluginsPath); err != nil {
		logMessage("ERROR", fmt.Sprintf("Error loading executor plugins: %v", err))
	} else {
		logMessage("INFO", "Loaded Executor plugins")
	}

	// Обработка манифеста
	if err := manifestHandler(*configPath); err != nil {
		logMessage("ERROR", fmt.Sprintf("Error loading manifest: %v", err))
	}

	logMessage("INFO", "RoLLer runner finish!")
	return nil
}
