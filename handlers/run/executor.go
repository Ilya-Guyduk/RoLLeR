package run

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
	// Создаем локацию
	pluginLocation, err := plugin.CreateLocation(locationData)
	logMessage("DEBUG", fmt.Sprintf("var:pluginLocation 'Location data for plugin: %+v'", pluginLocation))
	if err != nil || pluginLocation.Validate() != nil {
		return fmt.Errorf("ошибка с Location: %v", err)
	}

	// Создаем действие
	pluginAction, err := plugin.CreateAction(actionData)
	logMessage("DEBUG", fmt.Sprintf("var:pluginAction 'Action data for plugin: %+v", pluginAction))
	if err != nil {
		return fmt.Errorf("ошибка создания Action: %v", err)
	}

	// Выполняем подключение и действие
	if err := plugin.Execute(pluginLocation, pluginAction); err != nil { // Теперь передаем и location, и action
		return fmt.Errorf("ошибка выполнения действия через плагин: %v", err)
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
	logMessage("DEBUG", fmt.Sprintf("var:scriptLocationData 'Location data for Script: %+v'", scriptLocationData))
	scriptActionData := script.Actions
	logMessage("DEBUG", fmt.Sprintf("var:scriptActionData 'Action data for Script: %+v'", scriptActionData))

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
	logMessage("DEBUG", fmt.Sprintf("var:locationData 'Location data for Check: %+v'", locationData))
	actionData := check.Actions
	logMessage("DEBUG", fmt.Sprintf("var:actionData 'Action data for Check: %+v'", actionData))

	return executePluginAction(plugin, locationData, actionData)
}

// actionHandler выполняет указанные действия (например, pre-script или post-script).
func actionHandler(data interface{}, actionType string, atomicFlag *bool) error {

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	logMessage("DEBUG", fmt.Sprintf("ATOMIC: %v", atomicFlag))
	actionField := val.FieldByName(strings.Title(actionType))
	if !actionField.IsValid() {
		logMessage("DEBUG", fmt.Sprintf("Missing %s", actionType))
		return nil
	}
	logMessage("INFO", fmt.Sprintf("Executing %s...", actionType))
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
func checkHandler(data interface{}, checkType string, atomicFlag *bool) error {

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	checkField := val.FieldByName(strings.Title(checkType))
	if !checkField.IsValid() {
		logMessage("DEBUG", fmt.Sprintf("Missing %s", checkType))
		return nil
	}
	logMessage("INFO", fmt.Sprintf("Executing %s...", checkType))
	// Определяем тип проверки и выполняем ее.
	check := checkField.Interface()
	switch v := check.(type) {
	case Check:
		if err := executeCheck(v); err != nil {
			logMessage("ERROR", fmt.Sprintf("%s failed: %v", checkType, err))
			if *atomicFlag {
				return fmt.Errorf("%s failed: %v", checkType, err)
			}
		}
	default:
		if *atomicFlag {
			return fmt.Errorf("Unsupported check type for %s", checkType)
		} else {
			logMessage("DEBUG", fmt.Sprintf("Unsupported check type for %s", checkType))
		}
	}

	return nil
}
