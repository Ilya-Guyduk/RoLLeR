package plugin

import (
	"flag"
	"fmt"
	"os"
)

func HandleInstall(
	args []string,

) {
	installCmd := flag.NewFlagSet("install", flag.ExitOnError)
	installPackage := installCmd.String("package", "", "Package to install")
	installCmd.Parse(args)

	if *installPackage == "" {
		fmt.Println("Please specify a package to install using --package flag")
		os.Exit(1)
	}

	fmt.Printf("INFO", fmt.Sprintf("Installing package: %s", *installPackage))
	installPackageHandler(*installPackage)
}

func installPackageHandler(pkg string) {
	// Логика установки пакета
	fmt.Printf("Package %s installed successfully.\n", pkg)
}
