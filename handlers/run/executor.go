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

type Task struct {
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
func executePluginAction(plugin plugininterface.Connector, locationData, actionData map[string]interface{}, stageName string) error {
	// Создаем локацию
	pluginLocation, err := plugin.CreateLocation(locationData)
	logMessage("DEBUG", fmt.Sprintf("[%s] var:pluginLocation 'Location data for plugin: %+v'", stageName, pluginLocation))
	if err != nil || pluginLocation.Validate() != nil {
		return fmt.Errorf("ошибка с Location: %v", err)
	}

	// Создаем действие
	pluginAction, err := plugin.CreateAction(actionData)
	logMessage("DEBUG", fmt.Sprintf("[%s] var:pluginAction 'Action data for plugin: %+v", stageName, pluginAction))
	if err != nil {
		return fmt.Errorf("ошибка создания Action: %v", err)
	}

	// Выполняем подключение и действие
	if err := plugin.Execute(pluginLocation, pluginAction); err != nil { // Теперь передаем и location, и action
		return fmt.Errorf("ошибка выполнения действия через плагин: %v", err)
	}

	return nil
}

// executePluginItem универсально выполняет Script, Task или Check.
func executePluginItem(item interface{}, stageName string) error {
	// Найти плагин
	plugin, err := findPlugin(item)
	if err != nil {
		return err
	}

	// Логируем
	logMessage("INFO", fmt.Sprintf("[%s] Executing item: %+v", stageName, item))
	if DRY_RUN_FLAG {
		return nil
	}

	// Получаем данные локации и действия
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	locationData := val.FieldByName("Location").Interface().(map[string]interface{})
	actionData := val.FieldByName("Actions").Interface().(map[string]interface{})

	logMessage("DEBUG", fmt.Sprintf("[%s] var:locationData 'Location data: %+v'", stageName, locationData))
	logMessage("DEBUG", fmt.Sprintf("[%s] var:actionData 'Action data: %+v'", stageName, actionData))

	// Выполняем действие
	return executePluginAction(plugin, locationData, actionData, stageName)
}

// handler универсально выполняет actions и checks.
func handler(data interface{}, itemType string, atomicFlag *bool, stageName string) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	field := val.FieldByName(strings.Title(itemType))
	if !field.IsValid() || field.IsZero() {
		logMessage("DEBUG", fmt.Sprintf("[%s] Missing %s", stageName, itemType))
		return nil
	}

	logMessage("INFO", fmt.Sprintf("[%s] Executing %s...", stageName, itemType))
	item := field.Interface()

	// Универсальная обработка item (Script, Task, Check)
	if err := executePluginItem(item, stageName); err != nil {
		logMessage("ERROR", fmt.Sprintf("[%s] %s failed: %v", stageName, itemType, err))
		// Если это проверка (check), сразу завершить выполнение
		if strings.EqualFold(itemType, "preCheck") {
			logMessage("ERROR", fmt.Sprintf("[%s] Check failed, stopping execution.", stageName))
			os.Exit(1)
		}

		// Для других типов (например, actions), обработка зависит от atomicFlag
		if atomicFlag != nil && *atomicFlag {
			return fmt.Errorf("[%s] %s failed: %v", stageName, itemType, err)
		}
	}
	return nil
}
