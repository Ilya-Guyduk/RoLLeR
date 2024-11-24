package main

import (
	"fmt"
	"os"

	"github.com/Ilya-Guyduk/RoLLeR/handlers/inits"
	"github.com/Ilya-Guyduk/RoLLeR/handlers/plugin"
	"github.com/Ilya-Guyduk/RoLLeR/handlers/run"
)

const (
	MainBanner = `
  ___     _    _        ___        __   __   _ 
 | _ \___| |  | |   ___| _ \ __ __/  \ /  \ / |
 |   / _ \ |__| |__/ -_)   / \ V / () | () || |
 |_|_\___/____|____\___|_|_\  \_/ \__(_)__(_)_|    
 ==============================================                                                                                             
`
)

func main() {
	// Проверка наличия подкоманды
	if len(os.Args) < 2 {
		fmt.Println("Expected 'run', 'install', or 'search' subcommands")
		os.Exit(1)
	} else {
		fmt.Println(MainBanner)
	}

	// Обработка подкоманды
	switch os.Args[1] {
	case "run":
		run.HandleRun(
			os.Args[2:],
		)
	case "plugin":
		plugin.HandleInstall(
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
