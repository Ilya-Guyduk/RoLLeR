package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"plugin" // используем "plugin" для загрузки пользовательских библиотек
	"reflect"
	"strings"

	"github.com/Ilya-Guyduk/RoLLeR/plugininterface"
)

// KubernetesConfig реализует Connector для Kubernetes-подключений
type KubernetesConfig struct {
	Namespace string
}

type Location struct {
	KubeConfig KubernetesConfig `yaml:"KubeConfig"`
	PluginType string           `yaml:"plugin_type"` // определяет тип плагина
}

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

// PluginRegistry - глобальная карта для хранения зарегистрированных плагинов
var PluginRegistry = make(map[string]plugininterface.Connector)

// loadPlugins загружает все плагины из указанной папки
func loadExecutorPlugins(pluginsPath string) error {
	err := filepath.Walk(pluginsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".so") {
			// Загружаем файл плагина
			p, err := plugin.Open(path)
			if err != nil {
				return fmt.Errorf("ошибка загрузки плагина %s: %v", path, err)
			}

			// Ищем функцию "NewConnector" (должна возвращать Connector)
			symbol, err := p.Lookup("NewConnector")
			if err != nil {
				return fmt.Errorf("ошибка поиска функции NewConnector в плагине %s: %v", path, err)
			}

			connector, ok := symbol.(func() plugininterface.Connector)
			if !ok {
				return fmt.Errorf("NewConnector в плагине %s не реализует правильный интерфейс", path)
			}

			// Получаем имя плагина и регистрируем его
			pluginName := strings.TrimSuffix(info.Name(), ".so")
			PluginRegistry[pluginName] = connector()
			log.Printf("Загружен плагин: %s", pluginName)
		}
		return nil
	})
	return err
}

// findLocation находит и подключается к нужному плагину на основе его типа
func findLocation(data interface{}) (plugininterface.Connector, error) {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	locationField := val.FieldByName("Location")
	if !locationField.IsValid() || locationField.Kind() != reflect.Struct {
		return nil, errors.New("Location для подключения не найден")
	}

	pluginTypeField := locationField.FieldByName("PluginType")
	if !pluginTypeField.IsValid() || pluginTypeField.Kind() != reflect.String {
		return nil, errors.New("Тип подключения не указан в Location")
	}
	pluginType := pluginTypeField.String()
	logMessage("DEBUG", "pluginType %s", pluginType)

	plugin, ok := PluginRegistry[pluginType]
	if !ok {
		return nil, fmt.Errorf("плагин для типа '%s' не зарегистрирован", pluginType)
	}

	if err := plugin.Connect(); err != nil {
		return nil, fmt.Errorf("ошибка при подключении через плагин '%s': %v", pluginType, err)
	}
	return plugin, nil
}

// executePluginCommand выполняет пользовательское действие через плагин
func executePluginCommand(plugin plugininterface.Connector, action string) error {
	log.Printf("Выполнение действия: %s", action)
	return plugin.Execute(action)
}

// executeScript выполняет команду скрипта, если это Script.
func executeScript(script Script) error {
	// Здесь будет код для выполнения скрипта
	logMessage("INFO", fmt.Sprintf("Executing script: %+v", script))

	plugin, err := findLocation(script)

	// Если возникла критическая ошибка (не связанная с отсутствием одного из конфигов)
	if err != nil {
		return err
	}

	if plugin == nil {
		return fmt.Errorf("Plugin is empty")
	}
	return nil
}

// executeCheck теперь выполняет команды на основе конфигов.
func executeCheck(check Check) error {
	// Поиск локации выполнения чека
	plugin, err := findLocation(check)

	// Если возникла критическая ошибка (не связанная с отсутствием одного из конфигов)
	if err != nil {
		return err
	}

	if plugin == nil {
		return fmt.Errorf("Plugin is empty")
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
		//if hostConfig != nil {
		// Выполняем команду на хосте через SSH
		//	return executeHostCommand(hostConfig, check.Bash.User_script)
		//} else if kubeConfig != nil {
		// Выполняем команду в Kubernetes
		//	return executeK8sCommand(kubeConfig, check.Bash.User_script)
		//}
	} else if check.Run != "" {
		logMessage("INFO", fmt.Sprintf("Executing command: %s", check.Run))
		//cmdParts := strings.Fields(check.Run)
		//if hostConfig != nil {
		// Выполняем команду на хосте через SSH
		//	return executeHostCommand(hostConfig, cmdParts[0])
		//} else if kubeConfig != nil {
		// Выполняем команду в Kubernetes
		//	return executeK8sCommand(kubeConfig, cmdParts[0])
		//}
	}

	return nil
}
