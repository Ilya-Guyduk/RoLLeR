package plugin

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const githubRepoURL = "https://github.com/Ilya-Guyduk/RoLLeRHub"

const (
	hubBanner = `
  ___     _    _        ___ _  _      _    
 | _ \___| |  | |   ___| _ \ || |_  _| |__ 
 |   / _ \ |__| |__/ -_)   / __ | || | '_ \
 |_|_\___/____|____\___|_|_\_||_|\_,_|_.__/
===========================================
`
)

// HandlePluginCommand обрабатывает команды для управления плагинами
func HandlePluginCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Please specify a plugin command (e.g., install, search)")
		os.Exit(1)
	} else {
		fmt.Printf(hubBanner)
	}

	switch args[0] {
	case "install":
		handleInstall(args[1:])
	case "search":
		handleSearch(args[1:])
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		os.Exit(1)
	}
}

// handleInstall обрабатывает установку плагина
func handleInstall(args []string) {
	installCmd := flag.NewFlagSet("install", flag.ExitOnError)
	installPackage := installCmd.String("package", "", "Package to install")
	installCmd.Parse(args)

	if *installPackage == "" {
		fmt.Println("Please specify a package to install using --package flag")
		os.Exit(1)
	}

	fmt.Printf("INFO: Installing package: %s\n", *installPackage)
	installPackageHandler(*installPackage)
}

// handleSearch обрабатывает поиск плагина в репозитории
func handleSearch(args []string) {
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	pluginName := searchCmd.String("plugin", "", "Plugin to search for")
	searchCmd.Parse(args)

	if *pluginName == "" {
		fmt.Println("Please specify a plugin to search using --plugin flag")
		os.Exit(1)
	}

	fmt.Printf("INFO: Searching for plugin: %s\n", *pluginName)
	err := searchAndDownloadPlugin(*pluginName)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("INFO: Plugin %s downloaded successfully.\n", *pluginName)
}

// searchAndDownloadPlugin ищет плагин в репозитории и загружает его
func searchAndDownloadPlugin(pluginName string) error {
	pluginURL := fmt.Sprintf("%s/raw/main/plugins/%s.so", githubRepoURL, pluginName)
	resp, err := http.Get(pluginURL)
	if err != nil {
		return fmt.Errorf("failed to fetch plugin from URL %s: %v", pluginURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("plugin %s not found in repository", pluginName)
	}

	// Создаём директорию ./plugin, если её нет
	pluginDir := "./plugin"
	if err := os.MkdirAll(pluginDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create plugin directory: %v", err)
	}

	// Сохраняем плагин в файл
	pluginFilePath := filepath.Join(pluginDir, fmt.Sprintf("%s.so", pluginName))
	file, err := os.Create(pluginFilePath)
	if err != nil {
		return fmt.Errorf("failed to create plugin file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save plugin: %v", err)
	}

	return nil
}

// installPackageHandler выполняет логику установки пакета
func installPackageHandler(pkg string) {
	// Логика установки пакета
	fmt.Printf("Package %s installed successfully.\n", pkg)
}
