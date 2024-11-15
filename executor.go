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
}

type Script struct {
	Bash struct {
		User_script string `yaml:"script"`
	} `yaml:"bash"`
	Run      string   `yaml:"run"`
	Location Location `yaml:"location"`
}

type Check struct {
	PluginType string `yaml:"plugin"` // определяет тип плагина
	Actions    struct {
		User_script string `yaml:"script"`
	} `yaml:"bash"`
	Run      string   `yaml:"run"`
	Location Location `yaml:"location"`
}

// PluginRegistry - глобальная карта для хранения зарегистрированных плагинов
var PluginRegistry = make(map[string]plugininterface.Connector)

// loadExecutorPlugins загружает плагины и регистрирует их
func loadExecutorPlugins(pluginsPath string) error {
	return filepath.Walk(pluginsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".so") {
			// Загружаем файл плагина
			p, err := plugin.Open(path)
			if err != nil {
				return fmt.Errorf("ошибка загрузки плагина %s: %v", path, err)
			}

			// Ищем функцию NewConnector
			symbol, err := p.Lookup("NewConnector")
			if err != nil {
				return fmt.Errorf("ошибка поиска функции NewConnector в плагине %s: %v", path, err)
			}

			connectorFunc, ok := symbol.(func() plugininterface.Connector)
			if !ok {
				return fmt.Errorf("NewConnector в плагине %s не соответствует интерфейсу", path)
			}

			// Регистрируем плагин
			pluginName := strings.TrimSuffix(info.Name(), ".so")
			PluginRegistry[pluginName] = connectorFunc()
			logMessage("INFO", "Загружен плагин: %s", pluginName)
		}
		return nil
	})
}

// findLocation находит и подключается к нужному плагину на основе его типа
func findPlugin(data interface{}) (plugininterface.Connector, error) {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	pluginTypeField := val.FieldByName("PluginType")
	if !pluginTypeField.IsValid() || pluginTypeField.Kind() != reflect.Struct {
		return nil, errors.New("PluginType не указан!")
	}

	locationField := val.FieldByName("Location")
	if !locationField.IsValid() || locationField.Kind() != reflect.String {
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

	plugin, err := findPlugin(script)

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
	plugin, err := findPlugin(check)

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

// runPreActions выполняет pre-check и pre-script, если они есть.
func runPreActions(data interface{}) error {
	logMessage("DEBUG", "Search pre-actions...")

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	preCheckField := val.FieldByName("PreCheck")
	if !preCheckField.IsValid() || preCheckField.Kind() != reflect.Struct {
		logMessage("DEBUG", "Missing Pre_check")
	} else {
		if err := runAction(preCheckField, "pre_check", "PreCheck"); err != nil {
			return err
		}
	}

	locationField := val.FieldByName("preScript")
	if !locationField.IsValid() || locationField.Kind() != reflect.Struct {
		logMessage("DEBUG", "Missing pre_script")
	} else {
		if err := runAction(locationField, "pre_script", "preScript"); err != nil {
			return err
		}
	}

	return nil
}

// runPostActions выполняет post-check и post-script, если они есть.
func runPostActions(data interface{}) error {
	logMessage("DEBUG", "Starting post-actions...")

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	postCheckField := val.FieldByName("post-check")
	if !postCheckField.IsValid() || postCheckField.Kind() != reflect.Struct {
		logMessage("DEBUG", "Missing Post-check")
	} else {
		if err := runAction(postCheckField, "pre-check", "PreCheck"); err != nil {
			return err
		}
	}

	locationField := val.FieldByName("post-script")
	if !locationField.IsValid() || locationField.Kind() != reflect.Struct {
		logMessage("DEBUG", "Missing post-script")
	} else {
		if err := runAction(locationField, "pre-script", "preScript"); err != nil {
			return err
		}
	}

	return nil
}

// runAction выполняет проверку или скрипт, если они присутствуют.
func runAction(action interface{}, actionName, actionType string) error {
	switch v := action.(type) {
	case Check:
		// Если это Check, выполняем проверку
		logMessage("INFO", fmt.Sprintf("Running %s for stage", actionName))
		if err := executeCheck(v); err != nil {
			return fmt.Errorf("%s failed for stage: %v", actionType, err)
		}
	case Script:
		// Если это Script, выполняем скрипт
		logMessage("INFO", fmt.Sprintf("Running %s for stage", actionName))
		if err := executeScript(v); err != nil {
			return fmt.Errorf("%s failed for stage: %v", actionType, err)
		}
	default:
		// Если это не Check или Script, выводим debug сообщение
		logMessage("DEBUG", fmt.Sprintf("%s for stage is missing", actionType))
	}
	return nil
}

func executePluginAction(plugin plugininterface.Connector, locationData, actionData map[string]interface{}) error {
	// Получить Location из плагина
	location, err := plugin.CreateLocation(locationData)
	if err != nil {
		return fmt.Errorf("ошибка создания Location: %v", err)
	}

	// Проверить Location
	if err := location.Validate(); err != nil {
		return fmt.Errorf("некорректные данные Location: %v", err)
	}

	// Подключиться с использованием Location
	if err := plugin.Connect(location); err != nil {
		return fmt.Errorf("ошибка подключения через плагин: %v", err)
	}

	// Получить Action из плагина
	action, err := plugin.CreateAction(actionData)
	if err != nil {
		return fmt.Errorf("ошибка создания Action: %v", err)
	}

	// Выполнить Action
	if err := action.Execute(); err != nil {
		return fmt.Errorf("ошибка выполнения Action через плагин: %v", err)
	}

	return nil
}
