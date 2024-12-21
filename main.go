package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Ilya-Guyduk/RoLLeR/handlers/inits"
	"github.com/Ilya-Guyduk/RoLLeR/handlers/plugin"
	"github.com/Ilya-Guyduk/RoLLeR/handlers/run"
	"gopkg.in/yaml.v2"
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
	DEFAULT_PEI_VERSION = ""
)

var (
	DEFAULT_LOGGING_LEVEL     = `INFO`
	DEFAULT_LOGGING_FORMATTER = "default"
)

var (
	DEFAULT_CONFIG_PATH    = "./config.yml"
	DEFAULT_MIGRATION_PATH = "./migration.yml"
	DEFAULT_PLUGIN_DIR     = "./plugins"
	DEFAULT_REPO_DIR       = "./repos"
	DEFAULT_REPO           = "https://github.com/Ilya-Guyduk/RoLLeRHub/raw/main/index.json"
)

func initConfig(configPath string) (*RollerConfig, error) {

	// Чтение файла
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	// Разбор rollerConfig
	var rollerconfig RollerConfig
	err = yaml.Unmarshal(data, &rollerconfig)
	if err != nil {
		return nil, fmt.Errorf("error parsing YAML for rollerConfig: %v", err)
	}

	return &rollerconfig, nil
}

func initerCommandParser(args []string) error {
	return nil
}

func runnerCommandParser(args []string) error {

	if len(args) < 1 {
		log.Fatal("Please specify a plugin command (e.g., install, search)")
	}

	// Инициализация флагов
	runCmd, migrationPath, pluginsPath, config := setupRunnerFlags()
	if err := runCmd.Parse(args); err != nil {
		return fmt.Errorf(`Error parsing flags: %w`, err)
	}

	// Вызов логотипа
	fmt.Printf(MainBanner)
	rollerConfig, err := initConfig(*config)
	if err != nil {
		return err
	}

	setupLogging(rollerConfig.Global.Logging)

	logMessage("INFO", "RoLLeR Starting...")

	pc := &plugin.PluginController{}
	logMessage("DEBUG", "[PluginController] Creating PluginController")
	pc, pluginErr := pc.NewPluginController(*pluginsPath, DEFAULT_REPO_DIR, DEFAULT_REPO)
	if pluginErr != nil {
		logMessage("ERROR", "%s", pluginErr)
		return nil
	} else {
		logMessage("DEBUG", fmt.Sprintf("[PluginController] Version: %s, DefaultRepository: %s, LocalRepositoryPath: %s", pc.ControllerVersion, pc.DefaultRepository, pc.LocalRepositoryPath))
	}

	var migrationSet *run.MigrationSet
	// Инициализация MigrationSet
	logMessage("INFO", fmt.Sprintf("Creating MigrationSet: %s", migrationPath))
	migrationSet, migrationErr := migrationSet.NewMigrationSet(*migrationPath, pc, logMessage)
	if migrationErr != nil {
		logMessage("ERROR", "%s", migrationErr)
		return nil
	}

	// Каскадная валидация миграции
	logMessage("INFO", fmt.Sprintf("Start cascade validation"))
	validErr := migrationSet.CascadeValidation(*migrationSet, logMessage)
	if validErr != nil {
		logMessage("ERROR", "%s", validErr)
		return nil
	} else {
		logMessage("INFO", "[MigrationSet]>[Valid] Cascade validation finish!")
	}

	logMessage("INFO", fmt.Sprintf("Starting UpdateRelease"))
	updateErr := migrationSet.UpdateRelease(migrationSet, logMessage)
	if updateErr != nil {
		logMessage("ERROR", fmt.Sprintf("Error Update: %s", updateErr))
	}

	defer logMessage("INFO", "RoLLer runner finished")
	return nil
}

// setupFlags инициализирует флаги командной строки
func setupRunnerFlags() (*flag.FlagSet, *string, *string, *string) {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	migrationPath := runCmd.String("migration", DEFAULT_MIGRATION_PATH, "Path to the YAML migration file")
	pluginsPath := runCmd.String("pluginsPath", DEFAULT_PLUGIN_DIR, "Path to the plugins directory")
	config := runCmd.String("config", DEFAULT_CONFIG_PATH, "plugin to install")
	return runCmd, migrationPath, pluginsPath, config
}

func pluginCommandParser(args []string) error {
	if len(args) < 1 {
		log.Fatal("Please specify a plugin command (e.g., install, search)")
	}

	installCmd := flag.NewFlagSet("plugin", flag.ExitOnError)
	pluginName := installCmd.String("plugin", "", "plugin to install")
	config := installCmd.String("config", DEFAULT_CONFIG_PATH, "plugin to install")

	// Разбор флагов
	if err := installCmd.Parse(args); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *pluginName == "" {
		fmt.Println("Please specify a plugin to search using --plugin flag")
		os.Exit(1)
	}

	// Инициализация конфигурации
	rollerConfig, err := initConfig(*config)
	if err != nil {
		return err
	}

	setupLogging(rollerConfig.Global.Logging)

	logMessage("INFO", "RoLLeR PluginController")

	pc := &plugin.PluginController{}
	pc, pluginErr := pc.NewPluginController(rollerConfig.Global.Plugin.PluginPath, rollerConfig.Global.Plugin.PluginRepoPath, DEFAULT_REPO)
	if pluginErr != nil {
		return pluginErr
	}

	switch args[0] {
	case "install":
		fmt.Printf("INFO: Installing plugin: %s\n", *pluginName)
		installErr := pc.InstallPlugin(*pluginName)
		if installErr != nil {
			fmt.Println(installErr)
		}

	case "search":

		fmt.Printf("INFO: Searching for plugin: %s in %s\n", *pluginName, rollerConfig.Global.Plugin.DefaultRepo)

		PluginName, pluginVersion, pluginDescription, pluginURL, searchErr := pc.SearchPlugin(*pluginName, rollerConfig.Global.Plugin.DefaultRepo)
		if searchErr != nil {
			logMessage("ERROR", "%s", searchErr)
		} else {
			fmt.Printf("\nPlugin Found:\n")
			fmt.Printf("  Name: %s\n", PluginName)
			fmt.Printf("  Version: %s\n", pluginVersion)
			fmt.Printf("  Description: %s\n", pluginDescription)
			fmt.Printf("  URL: %s\n", pluginURL)
		}
	case "delete":

	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		os.Exit(1)
	}
	return nil
}

func main() {

	// Проверка наличия подкоманды
	if len(os.Args) < 2 {
		fmt.Println("Expected 'run', 'install', or 'search' subcommands")
		os.Exit(1)
	}
	// Обработка подкоманды
	switch os.Args[1] {
	case "run":
		runnerCommandParser(
			os.Args[2:],
		)
	case "plugin":

		pluginCommandParser(
			os.Args[2:],
		)
	case "init":
		inits.HandleInit(
			os.Args[2:],
		)
	default:
		fmt.Println("Expected 'run', 'init', or 'plugin' subcommands")
		os.Exit(1)
	}
}
