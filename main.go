package main

import (
	"flag"
	"fmt"
	"os"
)

var RUN_FLAG bool
var DRY_RUN_FLAG bool
var PLUGINS_PATH string

func main() {
	// Проверка наличия подкоманды
	if len(os.Args) < 2 {
		fmt.Println("Expected 'run', 'install', or 'search' subcommands")
		os.Exit(1)
	}

	// Объявляем подкоманды
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	dryRunFlag := runCmd.Bool("dryrun", false, "Dryrun the configuration steps")
	configPath := runCmd.String("config", "roller.yml", "Path to the YAML configuration file")
	logFlag := runCmd.String("loglevel", "INFO", "Set the log level (DEBUG, INFO, ERROR)")
	pluginsPath := runCmd.String("pluginsPath", "./plugins", "Plugins path")

	installCmd := flag.NewFlagSet("install", flag.ExitOnError)
	installPackage := installCmd.String("package", "", "Package to install")

	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	searchQuery := searchCmd.String("query", "", "Search query for packages")

	// Обработка подкоманд
	switch os.Args[1] {
	case "run":
		// Разбираем флаги для команды run
		runCmd.Parse(os.Args[2:])
		DRY_RUN_FLAG = *dryRunFlag
		PLUGINS_PATH = *pluginsPath
		setupLogging(*logFlag)

		logMessage("INFO", "RoLLer runner starting ...")

		// Загрузка плагинов исполнителя
		if err := loadExecutorPlugins(PLUGINS_PATH); err != nil {
			logMessage("ERROR", fmt.Sprintf("Error loading executor plugins: %v", err))
		} else {
			logMessage("INFO", "Loaded Executor plugins")
		}

		// Обработка манифеста
		if err := manifestHandler(*configPath); err != nil {
			logMessage("ERROR", fmt.Sprintf("Error loading manifest: %v", err))
		}

		defer logMessage("INFO", "RoLLer finish!")

	case "install":
		// Разбираем флаги для команды install
		installCmd.Parse(os.Args[2:])

		if *installPackage == "" {
			fmt.Println("Please specify a package to install using --package flag")
			os.Exit(1)
		}

		logMessage("INFO", fmt.Sprintf("Installing package: %s", *installPackage))
		// Здесь можно реализовать логику установки пакета
		installPackageHandler(*installPackage)

	case "search":
		// Разбираем флаги для команды search
		searchCmd.Parse(os.Args[2:])

		if *searchQuery == "" {
			fmt.Println("Please specify a search query using --query flag")
			os.Exit(1)
		}

		logMessage("INFO", fmt.Sprintf("Searching for: %s", *searchQuery))
		// Здесь можно реализовать логику поиска
		searchHandler(*searchQuery)

	default:
		fmt.Println("Expected 'run', 'install', or 'search' subcommands")
		os.Exit(1)
	}
}

func installPackageHandler(pkg string) {
	// Логика установки пакета
	fmt.Printf("Package %s installed successfully.\n", pkg)
}

func searchHandler(query string) {
	// Логика поиска
	fmt.Printf("Search results for '%s':\n", query)
	fmt.Println("- Package 1")
	fmt.Println("- Package 2")
	fmt.Println("- Package 3")
}
