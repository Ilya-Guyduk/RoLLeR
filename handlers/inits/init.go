package inits

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func createStandTemplate() string {
	return `stand:
  name: PROD
  desc: "Описание стенда"

  groups:
  - name: "MDM"
    k8s:
      namespace: "client-profile-prod"
      endpoints:
        release_name: "mdm-adapter"
    postgres:
      host:
      port:
      user:
      password:
`
}

func createReleaseTemplate() string {
	return `#
version: 
# Флаг атомарности
# В данном случае, распространяется на весь манифест
# Если что-либо завершено неуспешно в этом манифесте, 
# то весь манифест будет откатан до первоначалього состояния
atomic:
# Главная и начальная структура
stage:
# Уникальное имя этапа
# На него можно будет ссылаться при указании зависимостей
# Не должен содержать пробелы или кирилицу 
- name: 
  # Подробное описание действия
  # В данном случае, описание этапа
  desc:
  # Флаг атомарности
  # В данном случае, распространяется на весь этап
  # Если что-либо завершено неуспешно в этом этапе, 
  # то весь этап будет откатан до первоначалього состояния
  atomic:
  ...
`
}

const (
	IniterBanner = `	Initer v0.0.1
`
)

func createDirectories(basePath string) error {
	dirs := []string{
		basePath,
		filepath.Join(basePath, "stands"),
		filepath.Join(basePath, "release"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("ошибка при создании каталога %s: %v", dir, err)
		}
	}
	return nil
}

func createReleaseFile(basePath string) error {
	releaseFilePath := filepath.Join(basePath, "release", "release.yml")
	file, err := os.Create(releaseFilePath)
	if err != nil {
		return fmt.Errorf("ошибка при создании файла %s: %v", releaseFilePath, err)
	}
	defer file.Close()

	_, err = file.WriteString(createReleaseTemplate())
	if err != nil {
		return fmt.Errorf("ошибка при записи в файл %s: %v", releaseFilePath, err)
	}

	return nil
}

func createStandFile(basePath string) error {
	releaseFilePath := filepath.Join(basePath, "stands", "stands.yml")
	file, err := os.Create(releaseFilePath)
	if err != nil {
		return fmt.Errorf("ошибка при создании файла %s: %v", releaseFilePath, err)
	}
	defer file.Close()

	_, err = file.WriteString(createStandTemplate())
	if err != nil {
		return fmt.Errorf("ошибка при записи в файл %s: %v", releaseFilePath, err)
	}

	return nil
}

func HandleInit(args []string) {
	fmt.Println(IniterBanner)

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	subCommand := initCmd.String("type", "", "Type of init operation (e.g., 'release')")
	initCmd.Parse(args)

	if *subCommand == "" {
		fmt.Println("Please specify the init type using --type flag")
		os.Exit(1)
	}

	switch *subCommand {
	case "release":
		basePath := "roller"

		if err := createDirectories(basePath); err != nil {
			fmt.Printf("Ошибка при создании каталогов: %v\n", err)
			os.Exit(1)
		}

		if err := createReleaseFile(basePath); err != nil {
			fmt.Printf("Ошибка при создании release.yml: %v\n", err)
			os.Exit(1)
		}

		if err := createStandFile(basePath); err != nil {
			fmt.Printf("Ошибка при создании stand.yml: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Шаблон и структура каталога успешно созданы!")
	default:
		fmt.Printf("Неизвестный тип инициализации: %s\n", *subCommand)
		os.Exit(1)
	}
}
