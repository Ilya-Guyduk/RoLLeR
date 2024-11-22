package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"

	"github.com/Ilya-Guyduk/RoLLeR/plugininterface"
)

type Script struct {
	PluginType string                 `yaml:"plugin"`
	Actions    map[string]interface{} `yaml:"action"`
	Location   map[string]interface{} `yaml:"location"`
}

type Check struct {
	PluginType string                 `yaml:"plugin"`
	Actions    map[string]interface{} `yaml:"action"`
	Location   map[string]interface{} `yaml:"location"`
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
		logMessage("DEBUG", fmt.Sprintf("Loaded plugin: %s", info.Name()))
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
	if DRY_RUN_FLAG {
		return nil
	}

	scriptLocationData := script.Location
	scriptActionData := script.Actions

	return executePluginAction(plugin, scriptLocationData, scriptActionData)

}

// executeCheck выполняет Check, если это Check.
func executeCheck(check Check) error {
	plugin, err := findPlugin(check)
	if err != nil || plugin == nil {
		return err
	}

	logMessage("INFO", "Executing check script")

	if DRY_RUN_FLAG {
		return nil
	}

	locationData := check.Location
	actionData := check.Actions

	return executePluginAction(plugin, locationData, actionData)
}

// actionHandler выполняет указанные действия (например, pre-script или post-script).
func actionHandler(data interface{}, actionType string, atomicFlag bool) error {
	logMessage("INFO", fmt.Sprintf("Executing %s...", actionType))

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	actionField := val.FieldByName(strings.Title(actionType))
	if !actionField.IsValid() {
		logMessage("DEBUG", fmt.Sprintf("Missing %s", actionType))
		return nil
	}

	// Определяем тип действия и выполняем его.
	action := actionField.Interface()
	switch v := action.(type) {
	case Script:
		if err := executeScript(v); err != nil {
			return fmt.Errorf("%s failed: %v", actionType, err)
		}
	default:
		logMessage("DEBUG", fmt.Sprintf("Unsupported action type for %s", actionType))
	}

	return nil
}

// checkHandler выполняет проверки (например, pre-check или post-check).
func checkHandler(data interface{}, checkType string, atomicFlag bool) error {
	logMessage("INFO", fmt.Sprintf("Executing %s...", checkType))

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	checkField := val.FieldByName(strings.Title(checkType))
	if !checkField.IsValid() {
		logMessage("DEBUG", fmt.Sprintf("Missing %s", checkType))
		return nil
	}

	// Определяем тип проверки и выполняем ее.
	check := checkField.Interface()
	switch v := check.(type) {
	case Check:
		if err := executeCheck(v); err != nil {
			logMessage("ERROR", fmt.Sprintf("%s failed: %v", checkType, err))
			if atomicFlag {
				return fmt.Errorf("%s failed: %v", checkType, err)
			}
		}
	default:
		logMessage("DEBUG", fmt.Sprintf("Unsupported check type for %s", checkType))
	}

	return nil
}
