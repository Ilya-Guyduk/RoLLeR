package plugin

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// handleInstall обрабатывает установку плагина
func handleInstall(args []string) {
	installCmd := flag.NewFlagSet("install", flag.ExitOnError)
	installPlugin := installCmd.String("plugin", "", "plugin to install")
	installRepoURL := installCmd.String("repo", "", "plugin to install")
	installCmd.Parse(args)

	if *installRepoURL != "" {
		fmt.Printf("INFO: Installing repo: %s\n", *installRepoURL)
		err := installRepos(*installRepoURL)
		if err != nil {
			fmt.Printf("Error: %s", err)
		}
	} else if *installPlugin != "" {
		fmt.Printf("INFO: Installing package: %s\n", *installPlugin)
		err := searchAndDownloadPlugin(*installPlugin)
		if err != nil {
			fmt.Printf("Error: %s", err)
		}
	} else {
		fmt.Println("Please specify a plugin to install using --plugin flag")
		os.Exit(1)
	}

}

func installRepos(installRepoURL string) error {
	// Проверяем и исправляем URL
	fixedURL, err := validateAndFixURL(installRepoURL)
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
