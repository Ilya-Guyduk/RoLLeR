package plugin

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
	"time"

	v1 "github.com/Ilya-Guyduk/RoLLeR/pei/v1"
)

type PluginController struct {
	ControllerVersion      string
	ExecutorPluginRegistry map[string]v1.Executor
	PluginRepositoryMap    map[string]string
	LocalRepositoryPath    string
	DefaultRepository      string
}

func (pc *PluginController) InitPluginController(pluginsPath string, repoPath string, defaultRepo string) (*PluginController, error) {
	// Создайте новый экземпляр, если необходимо
	if pc == nil {
		pc = &PluginController{}
	}

	executorPluginRegistry, err := pc.loadExecutorPlugins(pluginsPath)
	if err != nil {
		return nil, err
	}
	// Создаем новый экземпляр MigrationSet с заполненными данными.
	newPC := &PluginController{
		ControllerVersion:      "0.0.1",
		ExecutorPluginRegistry: executorPluginRegistry,
		PluginRepositoryMap:    make(map[string]string),
		LocalRepositoryPath:    "",
		DefaultRepository:      indexFileURL,
	}
	return newPC, nil
}

func (pc *PluginController) FindExecutorPlugin(data interface{}) (v1.Executor, error) {

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	pluginTypeField := val.FieldByName("PluginType")
	if !pluginTypeField.IsValid() || pluginTypeField.Kind() != reflect.String {
		return nil, errors.New("неверное или отсутствующее поле PluginType")
	}

	pluginType := pluginTypeField.String()
	if pluginType == "" {
		return nil, errors.New("значение PluginType пустое")
	}

	executor, ok := pc.ExecutorPluginRegistry[pluginType]
	if !ok {
		return nil, fmt.Errorf("плагин для типа '%s' не найден", pluginType)
	}

	return executor, nil
}

func (pc *PluginController) InstallPlugin(pluginName string, repositoryURL string) error {

	err := pc.SearchPlugin(pluginName)
	if err != nil {
		return err
	}

	return nil
}

func (pc *PluginController) DeletePlugin(pluginName string) error {

	err := pc.SearchPlugin(pluginName)
	if err != nil {
		return err
	}

	return nil
}

func (pc *PluginController) SearchPlugin(pluginName string) error {

	return nil
}

func (pc *PluginController) AddRepo(repoJsonURL string) error {

	return nil
}

func (pc *PluginController) DeleteRepo(repoName string) error {

	return nil
}

func (pc *PluginController) loadExecutorPlugins(pluginsPath string) (map[string]v1.Executor, error) {
	executorPluginRegistry := make(map[string]v1.Executor)

	err := filepath.Walk(pluginsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("WARNING: Ошибка доступа к файлу %s: %v\n", path, err)
			return nil
		}

		// Пропускаем директории и файлы, не оканчивающиеся на ".so"
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".so") {
			return nil
		}

		// Загружаем плагин
		p, err := plugin.Open(path)
		if err != nil {
			fmt.Printf("WARNING: Ошибка загрузки плагина %s: %v\n", path, err)
			return nil
		}

		// Ищем функцию NewExecutor
		symbol, err := p.Lookup("NewExecutor")
		if err != nil {
			fmt.Printf("WARNING: Функция NewExecutor не найдена в плагине %s: %v\n", path, err)
			return nil
		}

		// Преобразуем символ в функцию
		executorFunc, ok := symbol.(func() v1.Executor)
		if !ok {
			fmt.Printf("WARNING: NewExecutor в плагине %s не соответствует интерфейсу Executor\n", path)
			return nil
		}

		// Создаем экземпляр плагина
		pluginInstance := executorFunc()
		pluginInfo, err := pluginInstance.GetInfo()
		if err != nil {
			fmt.Printf("WARNING: Ошибка получения информации о плагине %s: %v\n", path, err)
			return nil
		}

		// Добавляем плагин в реестр
		executorPluginRegistry[pluginInfo.Name] = pluginInstance
		fmt.Printf("INFO: Плагин %s успешно загружен.\n", pluginInfo.Name)

		return nil
	})

	// Если произошли ошибки обхода, логируем их, но возвращаем реестр
	if err != nil {
		fmt.Printf("WARNING: Ошибки при обходе плагинов в директории %s: %v\n", pluginsPath, err)
	}

	return executorPluginRegistry, nil
}

func (pc *PluginController) createAndCheckDir(dir string) error {

	// Создаём директорию ./plugin, если её нет
	targetdir := dir
	if err := os.MkdirAll(targetdir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create plugin directory: %v", err)
	}

	return nil
}

const githubRepoURL = "https://github.com/Ilya-Guyduk/RoLLeRHub"
const indexFileURL = githubRepoURL + "/raw/main/index.json"

type Plugin struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	URL          string   `json:"url"`
	Dependencies []string `json:"dependencies"`
}

type Index struct {
	Plugins []Plugin `json:"plugins"`
}

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

// handleSearch обрабатывает поиск плагина в index.json
func handleSearch(args []string) {
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	pluginName := searchCmd.String("plugin", "", "Plugin to search for")
	searchCmd.Parse(args)

	if *pluginName == "" {
		fmt.Println("Please specify a plugin to search using --plugin flag")
		os.Exit(1)
	}

	fmt.Printf("INFO: Searching for plugin: %s\n", *pluginName)
	err := searchPlugin(*pluginName)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}

// searchPlugin ищет плагин в index.json и выводит его информацию
func searchPlugin(pluginName string) error {
	localCacheDir := "./repos"
	localIndexPath := filepath.Join(localCacheDir, indexFileURL)

	// Убедимся, что каталог для кэша существует
	if err := os.MkdirAll(localCacheDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create local cache directory: %v", err)
	}

	// Проверяем, существует ли локальный файл index.json и его возраст
	fileInfo, err := os.Stat(localIndexPath)
	if err == nil {
		// Если файл существует и моложе 5 минут, используем его
		if time.Since(fileInfo.ModTime()) <= 5*time.Minute {
			fmt.Println("INFO: Using cached %s", indexFileURL)
			return searchInLocalIndex(localIndexPath, pluginName)
		}
		// Если файл старше 5 минут, удаляем его
		fmt.Println("INFO: Cached index.json is outdated, downloading a new version...")
		if err := os.Remove(localIndexPath); err != nil {
			return fmt.Errorf("failed to remove outdated index.json: %v", err)
		}
	} else if !os.IsNotExist(err) {
		// Если произошла ошибка при доступе к файлу, кроме его отсутствия
		return fmt.Errorf("failed to access local index.json: %v", err)
	}

	// Скачиваем свежую версию index.json
	if err := downloadIndexFile(localIndexPath); err != nil {
		return err
	}

	// Выполняем поиск в загруженном файле
	return searchInLocalIndex(localIndexPath, pluginName)
}

// downloadIndexFile загружает index.json и сохраняет его локально
func downloadIndexFile(localIndexPath string) error {
	resp, err := http.Get(indexFileURL)
	if err != nil {
		return fmt.Errorf("failed to download index.json: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download index.json: status code %d", resp.StatusCode)
	}

	// Сохраняем загруженный файл
	file, err := os.Create(localIndexPath)
	if err != nil {
		return fmt.Errorf("failed to create local index.json file: %v", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to save downloaded index.json: %v", err)
	}

	fmt.Println("INFO: index.json downloaded successfully")
	return nil
}

// searchInLocalIndex выполняет поиск плагина в локальном файле index.json
func searchInLocalIndex(localIndexPath, pluginName string) error {
	file, err := os.Open(localIndexPath)
	if err != nil {
		return fmt.Errorf("failed to open local index.json: %v", err)
	}
	defer file.Close()

	var index Index
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&index); err != nil {
		return fmt.Errorf("failed to decode local index.json: %v", err)
	}

	// Поиск плагина
	for _, plugin := range index.Plugins {
		if plugin.Name == pluginName {
			// Вывод информации о найденном плагине
			fmt.Printf("\nPlugin Found:\n")
			fmt.Printf("  Name: %s\n", plugin.Name)
			fmt.Printf("  Version: %s\n", plugin.Version)
			fmt.Printf("  Description: %s\n", plugin.Description)
			fmt.Printf("  URL: %s\n", plugin.URL)
			if len(plugin.Dependencies) > 0 {
				fmt.Printf("  Dependencies: %s\n", plugin.Dependencies)
			} else {
				fmt.Printf("  Dependencies: None\n")
			}
			return nil
		}
	}

	return errors.New("plugin not found in index.json")
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

	// Проверяем тип контента
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/octet-stream" {
		return errors.New("downloaded file is not a valid binary plugin")
	}

	// Создаём директорию ./plugin, если её нет
	pluginDir := "./plugins"
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

	// Проверяем размер файла
	info, err := file.Stat()
	if err != nil || info.Size() == 0 {
		return errors.New("downloaded plugin is empty or corrupted")
	}

	return nil
}
