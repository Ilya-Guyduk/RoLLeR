package main

import (
	"fmt"
	"os"

	"github.com/Ilya-Guyduk/RoLLeR/handlers/inits"
	"github.com/Ilya-Guyduk/RoLLeR/handlers/plugin"
	"github.com/Ilya-Guyduk/RoLLeR/handlers/run"
)

func main() {
	// Проверка наличия подкоманды
	if len(os.Args) < 2 {
		fmt.Println("Expected 'run', 'install', or 'search' subcommands")
		os.Exit(1)
	}
	// Обработка подкоманды
	switch os.Args[1] {
	case "run":

		run.HandleRun(
			os.Args[2:],
		)
	case "plugin":
		plugin.HandlePluginCommand(
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
