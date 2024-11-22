package main

import (
	"fmt"
	"os"

	"github.com/Ilya-Guyduk/RoLLeR/handlers/install"
	"github.com/Ilya-Guyduk/RoLLeR/handlers/run"
	"github.com/Ilya-Guyduk/RoLLeR/handlers/search"
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
	case "install":
		install.HandleInstall(
			os.Args[2:],
		)
	case "search":
		search.HandleSearch(
			os.Args[2:],
		)
	default:
		fmt.Println("Expected 'run', 'install', or 'search' subcommands")
		os.Exit(1)
	}
}
