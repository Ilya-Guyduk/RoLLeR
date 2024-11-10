package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type Script struct {
	Bash struct {
		User_script string `yaml:"script"`
	} `yaml:"bash"`
	Run      string   `yaml:"run"`
	Location Location `yaml:"location"`
}

type Check struct {
	Bash struct {
		User_script string `yaml:"script"`
	} `yaml:"bash"`
	Run      string   `yaml:"run"`
	Location Location `yaml:"location"`
}

// executeScript выполняет команду скрипта, если это Script.
func executeScript(script Script) error {
	// Здесь будет код для выполнения скрипта
	logMessage("INFO", fmt.Sprintf("Executing script: %+v", script))
	return nil
}

// executeCheck теперь выполняет команды на основе конфигов.
func executeCheck(check Check) error {
	// Поиск локации выполнения чека
	hostConfig, kubeConfig, err := findLocation(check)

	// Если возникла критическая ошибка (не связанная с отсутствием одного из конфигов)
	if err != nil {
		return fmt.Errorf("Check failed: %v", err)
	}

	// Проверка и дальнейшая работа с найденными конфигурациями.
	if hostConfig != nil {
		logMessage("DEBUG", fmt.Sprintf("HostConfig найден: %+v", *hostConfig))
	} else {
		logMessage("DEBUG", "HostConfig не найден, использование значений по умолчанию")
	}

	if kubeConfig != nil {
		logMessage("DEBUG", fmt.Sprintf("KubernetesConfig найден: %+v", *kubeConfig))
	} else {
		logMessage("DEBUG", "KubernetesConfig не найден")
	}

	// Если DRY_RUN_FLAG установлен, только логируем, не выполняя команду
	if DRY_RUN_FLAG {
		if check.Bash.User_script != "" {
			logMessage("INFO", fmt.Sprintf("Executing check script: %s", check.Bash.User_script))
		} else if check.Run != "" {
			logMessage("INFO", fmt.Sprintf("Executing command: %s", check.Run))
		}
		return nil // Не выполняем команду, только логируем
	}

	// Если флаг не установлен, выполняем команду как обычно
	if check.Bash.User_script != "" {
		logMessage("INFO", fmt.Sprintf("Executing check script: %s", check.Bash.User_script))
		if hostConfig != nil {
			// Выполняем команду на хосте через SSH
			return executeHostCommand(hostConfig, check.Bash.User_script)
		} else if kubeConfig != nil {
			// Выполняем команду в Kubernetes
			return executeK8sCommand(kubeConfig, check.Bash.User_script)
		}
	} else if check.Run != "" {
		logMessage("INFO", fmt.Sprintf("Executing command: %s", check.Run))
		cmdParts := strings.Fields(check.Run)
		if hostConfig != nil {
			// Выполняем команду на хосте через SSH
			return executeHostCommand(hostConfig, cmdParts[0])
		} else if kubeConfig != nil {
			// Выполняем команду в Kubernetes
			return executeK8sCommand(kubeConfig, cmdParts[0])
		}
	}

	return nil
}

// executeHostCommand выполняет команду на удаленном хосте через SSH.
// executeHostCommand выполняет команду на удаленном хосте через SSH.
func executeHostCommand(config *HostConfig, command string) error {
	// Создаем SSH-конфигурацию
	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password), // Используем пароль для аутентификации
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Используем insecure callback для игнорирования проверки хоста
		Timeout:         10 * time.Second,            // Таймаут для подключения
	}

	// Формируем строку адреса (IP:порт)
	address := fmt.Sprintf("%s:%d", config.Address, config.Port)

	// Подключаемся к SSH-серверу
	client, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to host %s: %v", address, err)
	}
	defer client.Close() // Закрываем соединение по завершению работы

	// Открываем новый сессионный канал
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close() // Закрываем сессию после выполнения команды

	// Устанавливаем вывод команды в терминал
	session.Stdout = log.Writer()
	session.Stderr = log.Writer()

	// Подготовка команды
	cmd := strings.Join([]string{command}, " ")

	// Выполняем команду
	logMessage("INFO", fmt.Sprintf("Executing command '%s' on host %s:%d", command, config.Address, config.Port))
	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to execute command on host %s: %v", config.Address, err)
	}

	// Возвращаем nil, если команда выполнена успешно
	return nil
}

// executeK8sCommand выполняет команду в Kubernetes, например, с использованием kubectl.
func executeK8sCommand(config *KubernetesConfig, command string) error {
	// Для выполнения команд в Kubernetes можно использовать kubectl через командную строку
	// или использовать Kubernetes API. Здесь пример использования командной строки.
	logMessage("INFO", fmt.Sprintf("Executing command '%s' in Kubernetes namespace '%s'", command, config.Namespace))
	// Возвращаем nil, так как это просто пример
	return nil
}
