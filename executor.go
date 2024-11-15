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

type Script struct {
	PluginType string `yaml:"plugin"` // определяет тип плагина
	Actions    struct {
		User_script string `yaml:"action"`
	} `yaml:"bash"`
	Run      string   `yaml:"run"`
	Location struct{} `yaml:"location"`
}

type Check struct {
	PluginType string `yaml:"plugin"` // определяет тип плагина
	Actions    struct {
		User_script string `yaml:"action"`
	} `yaml:"bash"`
	Run      string                 `yaml:"run"`
	Location map[string]interface{} `yaml:"location"`
}

// PluginRegistry - глобальная карта для хранения зарегистрированных плагинов
var PluginRegistry = make(map[string]plugininterface.Connector)

// loadExecutorPlugins загружает все плагины из указанной директории
func loadExecutorPlugins(pluginsPath string) error {
	return filepath.Walk(pluginsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".so") {
			return err
		}

		p, err := plugin.Open(path)
		if err != nil {
			return fmt.Errorf("ошибка загрузки плагина %s: %v", path, err)
		}

		symbol, err := p.Lookup("NewConnector")
		if err != nil {
			return fmt.Errorf("ошибка поиска функции NewConnector в плагине %s: %v", path, err)
		}

		connectorFunc, ok := symbol.(func() plugininterface.Connector)
		if !ok {
			return fmt.Errorf("NewConnector в плагине %s не соответствует интерфейсу", path)
		}

		PluginRegistry[strings.TrimSuffix(info.Name(), ".so")] = connectorFunc()
		log.Printf("Загружен плагин: %s", info.Name())
		return nil
	})
}

// findPlugin находит плагин, используя тип плагина из структуры
func findPlugin(data interface{}) (plugininterface.Connector, error) {
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

	plugin, ok := PluginRegistry[pluginType]
	if !ok {
		return nil, fmt.Errorf("плагин для типа '%s' не найден", pluginType)
	}

	return plugin, nil
}

// executePluginAction выполняет действия с плагином
func executePluginAction(plugin plugininterface.Connector, locationData, actionData map[string]interface{}) error {
	location, err := plugin.CreateLocation(locationData)
	if err != nil || location.Validate() != nil {
		return fmt.Errorf("ошибка с Location: %v", err)
	}

	if err := plugin.Connect(location); err != nil {
		return fmt.Errorf("ошибка подключения через плагин: %v", err)
	}

	action, err := plugin.CreateAction(actionData)
	if err != nil {
		return fmt.Errorf("ошибка создания Action: %v", err)
	}

	if err := plugin.Execute(action); err != nil {
		return fmt.Errorf("ошибка выполнения Action: %v", err)
	}

	return nil
}

// executeScript выполняет скрипт, если это Script.
func executeScript(script Script) error {
	plugin, err := findPlugin(script)
	if err != nil || plugin == nil {
		return err
	}

	logMessage("INFO", fmt.Sprintf("Executing script: %+v", script))
	return nil
}

// executeCheck выполняет Check, если это Check.
func executeCheck(check Check) error {
	plugin, err := findPlugin(check)
	if err != nil || plugin == nil {
		return err
	}

	logMessage("INFO", "Executing check script:")

	if DRY_RUN_FLAG {
		return nil
	}

	locationData := check.Location
	actionData := map[string]interface{}{"action": check.Actions.User_script, "run": check.Run}

	return executePluginAction(plugin, locationData, actionData)
}

// runAction выполняет проверку или скрипт.
func runAction(action interface{}, actionName string) error {
	switch v := action.(type) {
	case Check:
		if err := executeCheck(v); err != nil {
			return fmt.Errorf("%s failed: %v", actionName, err)
		}
	case Script:
		if err := executeScript(v); err != nil {
			return fmt.Errorf("%s failed: %v", actionName, err)
		}
	default:
		logMessage("DEBUG", fmt.Sprintf("%s is missing", actionName))
	}
	return nil
}

// runPreActions и runPostActions используют общую логику для обработки pre и post действий.
func runActions(data interface{}, actionName string) error {
	logMessage("DEBUG", fmt.Sprintf("Searching %s...", actionName))

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	actionField := val.FieldByName(actionName)
	if !actionField.IsValid() || actionField.Kind() != reflect.Struct {
		logMessage("DEBUG", fmt.Sprintf("Missing %s", actionName))
		return nil
	}

	return runAction(actionField.Interface(), actionName)
}

// runPreActions и runPostActions теперь используют runActions.
func runPreActions(data interface{}) error {
	if err := runActions(data, "PreCheck"); err != nil {
		return err
	}

	if err := runActions(data, "preScript"); err != nil {
		return err
	}

	return nil
}

func runPostActions(data interface{}) error {
	if err := runActions(data, "post-check"); err != nil {
		return err
	}

	if err := runActions(data, "post-script"); err != nil {
		return err
	}

	return nil
}
