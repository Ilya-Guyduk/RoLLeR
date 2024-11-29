package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
)

// PluginRegistry - глобальная карта для хранения зарегистрированных плагинов
var PluginRegistry = make(map[string]v1.Executor)

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

		symbol, err := p.Lookup("NewExecutor")
		if err != nil {
			return fmt.Errorf("ошибка поиска функции NewExecutor в плагине %s: %v", path, err)
		}

		executorFunc, ok := symbol.(func() v1.Executor)
		if !ok {
			return fmt.Errorf("NewExecutor в плагине %s не соответствует интерфейсу Executor", path)
		}

		pluginInstance := executorFunc()
		info, err := pluginInstance.GetInfo()
		if err != nil {
			return fmt.Errorf("ошибка получения информации о плагине %s: %v", path, err)
		}

		PluginRegistry[info.Name] = pluginInstance
		logMessage("DEBUG", fmt.Sprintf("Loaded plugin: %s", info.Name))
		return nil
	})
}

// findPlugin находит плагин, используя тип плагина из структуры
func findPlugin(data interface{}) (v1.Executor, error) {
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

	executor, ok := PluginRegistry[pluginType]
	if !ok {
		return nil, fmt.Errorf("плагин для типа '%s' не найден", pluginType)
	}

	return executor, nil
}

// executeExecutorItem универсально выполняет действие (Script, Task, Check) через Executor.
func executeExecutorItem(item interface{}, stageName string) error {
	// Найти плагин
	executor, err := findPlugin(item)
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

	// Проверяем данные
	if err := executor.ValidateYAML(actionData); err != nil {
		return fmt.Errorf("ошибка валидации данных: %v", err)
	}

	// Выполняем действие
	ctx := context.TODO() // Контекст можно адаптировать под требования
	if err := executor.Execute(ctx); err != nil {
		status := executor.GetStatus()
		logMessage("ERROR", fmt.Sprintf("[%s] Action failed: %s", stageName, status.Message))
		return err
	}

	// Получаем статус и логируем результат
	status := executor.GetStatus()
	logMessage("INFO", fmt.Sprintf("[%s] Action completed: %s", stageName, status.Message))
	return nil
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
	if err := executeExecutorItem(item, stageName); err != nil {
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
