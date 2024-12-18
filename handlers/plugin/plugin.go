package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
	"time"

	v1 "github.com/Ilya-Guyduk/RoLLeR/pei/v1"
)

const (
	hubBanner = `
  ___     _    _        ___ _  _      _    
 | _ \___| |  | |   ___| _ \ || |_  _| |__ 
 |   / _ \ |__| |__/ -_)   / __ | || | '_ \
 |_|_\___/____|____\___|_|_\_||_|\_,_|_.__/
===========================================
`
)

var (
	ROOT_INDEX_FILE_NAME = "_index.json"
)

type Plugin struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	URL          string   `json:"url"`
	Dependencies []string `json:"dependencies"`
	Hash         string   `json:"hash"`
}

type Repo struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	URL            string `json:"url"`
	LocalIndexFile string `json:"localIndex"`
}

type Index struct {
	Plugins []Plugin `json:"plugins"`
}

type RepoIndex struct {
	Repos []Repo `json:"repos"`
}

type PluginController struct {
	ControllerVersion      string
	ExecutorPluginRegistry map[string]v1.Executor
	PluginRepositoryMap    map[string]string
	LocalRepositoryPath    string
	RootRepositoryIndex    string
	DefaultRepository      string
}

func (pc *PluginController) NewPluginController(pluginsPath string, repoPath string, defaultRepo string) (*PluginController, error) {
	// Создайте новый экземпляр, если необходимо
	if pc == nil {
		pc = &PluginController{}
	}

	executorPluginRegistry, err := pc.loadExecutorPlugins(pluginsPath)
	if err != nil {
		return nil, err
	}
	rootIndexPath := filepath.Join(repoPath, ROOT_INDEX_FILE_NAME)
	// Создаем новый экземпляр MigrationSet с заполненными данными.
	newPC := &PluginController{
		ControllerVersion:      "0.0.1",
		ExecutorPluginRegistry: executorPluginRegistry,
		PluginRepositoryMap:    make(map[string]string),
		LocalRepositoryPath:    repoPath,
		RootRepositoryIndex:    rootIndexPath,
		DefaultRepository:      defaultRepo,
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

func (pc *PluginController) InstallPlugin(pluginName string) error {

	_, _, _, pluginURL, err := pc.SearchPlugin(pluginName, pc.DefaultRepository)
	if err != nil {
		return err
	}

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

func (pc *PluginController) DeletePlugin(pluginName string) error {

	return nil
}

func (pc *PluginController) SearchPlugin(pluginName string, repository string) (string, string, string, string, error) {

	fmt.Printf("PluginController: Searching for plugin: %s\n", pluginName)

	// Убедимся, что каталог для кэша существует

	if err := os.MkdirAll(pc.LocalRepositoryPath, os.ModePerm); err != nil {
		return "", "", "", "", fmt.Errorf("failed to create local cache directory: %v", err)
	}

	fmt.Printf("PluginController: Check %s\n", pc.RootRepositoryIndex)
	// Проверяем, существует ли локальный файл '_index.json'
	_, statErr := os.Stat(pc.RootRepositoryIndex)
	if statErr == nil {

		file, err := os.Open(pc.RootRepositoryIndex)
		if err != nil {
			return "", "", "", "", fmt.Errorf("failed to open local '_index.json': %v", err)
		}
		defer file.Close()

		fmt.Printf("PluginController: Decode %s\n", pc.RootRepositoryIndex)
		var repoIndex RepoIndex
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&repoIndex); err != nil {
			return "", "", "", "", fmt.Errorf("failed to decode local '_index.json': %v", err)
		}

		// Поиск репозитория
		for _, Repo := range repoIndex.Repos {
			if Repo.Name == repository {

				// Проверяем, существует ли локальный индекс репозитория '*.json' и его возраст
				fileInfo, err := os.Stat(Repo.LocalIndexFile)
				if err == nil {
					// Если файл существует и моложе 5 минут, используем его
					if time.Since(fileInfo.ModTime()) <= 5*time.Minute {
						fmt.Println("INFO: Using cached %s", Repo.LocalIndexFile)
						return searchInLocalIndex(Repo.LocalIndexFile, pluginName)
					}
					// Если файл старше 5 минут, удаляем его
					fmt.Println("INFO: Cached index.json is outdated, downloading a new version...")
					err := pc.UpdateRepoFile(Repo.Name)
					if err != nil {
						return "", "", "", "", err
					}

					// Выполняем поиск в загруженном файле
					return searchInLocalIndex(Repo.LocalIndexFile, pluginName)

				} else {
					// Скачиваем свежую версию index.json
					if err := pc.downloadIndexFile(Repo.URL, repository); err != nil {
						return "", "", "", "", err
					}
					// Выполняем поиск в загруженном файле
					return searchInLocalIndex(Repo.LocalIndexFile, pluginName)
				}

			}
		}

	}
	// Если произошла ошибка при доступе к файлу, кроме его отсутствия
	return "", "", "", "", fmt.Errorf("failed to access local _index.json: %v", statErr)
}

func (pc *PluginController) AddRepo(repoJsonURL string) error {

	// Проверяем и исправляем URL
	fixedURL, err := validateAndFixURL(repoJsonURL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %v", err)
	}

	// Загружаем JSON-файл
	resp, err := http.Get(fixedURL)
	if err != nil {
		return fmt.Errorf("failed to fetch repository file from URL %s: %v", fixedURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("repository file not found at URL %s: status %d", fixedURL, resp.StatusCode)
	}

	// Читаем содержимое файла
	repoData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read repository file: %v", err)
	} else {
		fmt.Printf("repoData: %s", repoData)
	}

	// Парсим JSON
	var repoInfo struct {
		RepoName    string `json:"repoName"`
		RepoLogo    string `json:"repoLogo"`
		Description string `json:"description"`
		IndexURL    string `json:"indexURL"`
	}
	if err := json.Unmarshal(repoData, &repoInfo); err != nil {
		return fmt.Errorf("failed to parse repository JSON: %v", err)
	}

	// Проверяем и исправляем URL для index.json
	indexURL, err := validateAndFixURL(repoInfo.IndexURL)
	if err != nil {
		return fmt.Errorf("invalid index URL in repository JSON: %v", err)
	} else {
		fmt.Printf("Index URL: %s", indexURL)
	}

	// Скачиваем index.json
	resp, err = http.Get(indexURL)
	if err != nil {
		return fmt.Errorf("failed to fetch index.json from URL %s: %v", indexURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("index.json not found at URL %s: status %d", indexURL, resp.StatusCode)
	}

	// Убедимся, что директория ./repos существует
	repoDir := "./repos"
	if err := os.MkdirAll(repoDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create repos directory: %v", err)
	}

	// Сохраняем index.json в файл
	indexFilePath := filepath.Join(repoDir, repoInfo.RepoName+".json")
	file, err := os.Create(indexFilePath)
	if err != nil {
		return fmt.Errorf("failed to create index.json file: %v", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to save index.json: %v", err)
	}

	fmt.Printf("INFO: Repository '%s' successfully installed.\n", repoInfo.RepoName)
	return nil
}

func (pc *PluginController) DeleteRepo(repoName string) error {

	return nil
}

func (pc *PluginController) UpdateRepoFile(repositoryNane string) error {

	_, err := os.Stat(pc.RootRepositoryIndex)
	if err == nil {

		file, err := os.Open(pc.RootRepositoryIndex)
		if err != nil {
			return fmt.Errorf("failed to open local '_index.json': %v", err)
		}
		defer file.Close()

		var repoIndex RepoIndex
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&repoIndex); err != nil {
			return fmt.Errorf("failed to decode local '_index.json': %v", err)
		}
		for _, Repo := range repoIndex.Repos {
			if Repo.Name == repositoryNane {

				if err := os.Remove(Repo.LocalIndexFile); err != nil {
					return fmt.Errorf("failed to remove outdated index.json: %v", err)
				}

				// Скачиваем свежую версию index.json
				if err := pc.downloadIndexFile(Repo.URL, repositoryNane); err != nil {
					return err
				}
				return nil

			}

		}
	}
	return nil

}

func (pc *PluginController) loadExecutorPlugins(pluginsPath string) (map[string]v1.Executor, error) {
	// Инициализация карты, если она не была инициализирована
	if pc.ExecutorPluginRegistry == nil {
		pc.ExecutorPluginRegistry = make(map[string]v1.Executor)
	}

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
		newExecutorFunc, ok := symbol.(func() v1.Executor)
		if !ok {
			fmt.Printf("WARNING: NewExecutor в плагине %s не соответствует интерфейсу Executor\n", path)
			return nil
		}

		// Создаем экземпляр плагина
		executorInstance := newExecutorFunc()

		// Получаем информацию о плагине
		pluginInfo, err := executorInstance.GetInfo()
		if err != nil {
			fmt.Printf("WARNING: Ошибка получения информации о плагине %s: %v\n", path, err)
			return nil
		}

		// Добавляем плагин в реестр
		pc.ExecutorPluginRegistry[pluginInfo.Name] = executorInstance
		fmt.Printf("Плагин %s успешно загружен.\n", pluginInfo.Name)

		return nil
	})

	// Если произошли ошибки обхода, логируем их, но возвращаем реестр
	if err != nil {
		fmt.Printf("WARNING: Ошибки при обходе плагинов в директории %s: %v\n", pluginsPath, err)
	}

	return pc.ExecutorPluginRegistry, nil
}

func (pc *PluginController) createAndCheckDir(dir string) error {

	// Создаём директорию ./plugin, если её нет
	targetdir := dir
	if err := os.MkdirAll(targetdir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create plugin directory: %v", err)
	}

	return nil
}

// downloadIndexFile загружает index.json и сохраняет его локально
func (pc *PluginController) downloadIndexFile(indexFileURL string, name string) error {
	resp, err := http.Get(indexFileURL)
	if err != nil {
		return fmt.Errorf("failed to download index.json: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download index.json: status code %d", resp.StatusCode)
	}

	// Сохраняем загруженный файл
	file, err := os.Create(pc.LocalRepositoryPath + "/" + name + ".json")
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
func searchInLocalIndex(localIndexPath, pluginName string) (string, string, string, string, error) {
	file, err := os.Open(localIndexPath)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to open local index.json: %v", err)
	}
	defer file.Close()

	var index Index
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&index); err != nil {
		return "", "", "", "", fmt.Errorf("failed to decode local index.json: %v", err)
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
			return plugin.Name, plugin.Version, plugin.Description, plugin.URL, nil
		}
	}

	return "", "", "", "", errors.New("plugin not found in index.json")
}

// validateAndFixURL проверяет URL и исправляет его при необходимости
func validateAndFixURL(inputURL string) (string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %v", err)
	}

	// Если в URL отсутствует схема, добавляем ее
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	// Если в URL отсутствует хост, считаем URL некорректным
	if parsedURL.Host == "" {
		return "", fmt.Errorf("URL missing host: %s", inputURL)
	}

	return parsedURL.String(), nil
}
