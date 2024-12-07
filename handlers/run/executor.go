package run

import (
	"context"
	"fmt"

	//"plugin"
	"reflect"

	v1 "github.com/Ilya-Guyduk/RoLLeR/pei/v1"
)

type Script struct {
	Name       string                 `yaml:"name"`
	PluginType string                 `yaml:"plugin"`
	Actions    map[string]interface{} `yaml:"action"`
	Component  string                 `yaml:"component"`
}

func (s *Script) CheckValideData(script Script) error {

	executor, ok := PluginRegistry[script.PluginType]
	if !ok {
		return fmt.Errorf("'script.Plugin' плагин для типа '%s' не найден", script.PluginType)
	}
	// Проверяем данные
	if err := executor.ValidateYAML(script.Actions); err != nil {
		return fmt.Errorf("ошибка валидации данных: %v", err)
	}
	return nil
}

func (s *Script) ExecScript(item interface{}, stageName string) error {

	return nil
}

type Check struct {
	Set        *MigrationSet
	Name       string                 `yaml:"name"`
	PluginType string                 `yaml:"plugin"`
	Actions    map[string]interface{} `yaml:"action"`
	Component  map[string]interface{} `yaml:"component"`
}

func (c *Check) CheckValideData(check Check) error {

	pc := c.Set.PluginController

	executor, ok := pc.ExecutorPluginRegistry[check.PluginType]
	if !ok {
		return fmt.Errorf("'component.Plugin' плагин для типа '%s' не найден", check.PluginType)
	}
	// Проверяем данные
	if err := executor.ValidateYAML(check.Actions); err != nil {
		return fmt.Errorf("ошибка валидации данных: %v", err)
	}

	component, err := c.Set.StandsFile.FindComponent(c.Component)
	if err != nil {
		return err
	}
	componentErr := executor.ValidateYAML(component)
	if componentErr != nil {
		return err
	}

	return nil
}

func (c *Check) ExecCheck(item interface{}, stageName string) error {

	// Найти плагин
	executor, err := c.Set.PluginController.FindExecutorPlugin(c.PluginType)
	if err != nil {
		return err
	}

	// Получаем данные локации и действия
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	componentData := val.FieldByName("Component").Interface().(map[string]interface{})
	actionData := val.FieldByName("Actions").Interface().(map[string]interface{})

	// Проверяем данные
	if err := executor.ValidateYAML(actionData); err != nil {
		return fmt.Errorf("ошибка валидации данных: %v", err)
	}

	// Проверяем данные
	if err := executor.ValidateYAML(componentData); err != nil {
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

type Task struct {
	Set        *MigrationSet
	PluginType string                 `yaml:"plugin"`
	Actions    map[string]interface{} `yaml:"action"`
	Component  string                 `yaml:"component"`
}

func (t *Task) ExecTask(item interface{}, stageName string) error {

	// Найти плагин
	executor, err := t.Set.PluginController.FindExecutorPlugin(t.PluginType)
	if err != nil {
		return err
	}

	// Получаем данные локации и действия
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	componentData := val.FieldByName("Component").Interface().(map[string]interface{})
	actionData := val.FieldByName("Actions").Interface().(map[string]interface{})

	// Проверяем данные
	if err := executor.ValidateYAML(actionData); err != nil {
		return fmt.Errorf("ошибка валидации данных: %v", err)
	}

	// Проверяем данные
	if err := executor.ValidateYAML(componentData); err != nil {
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

func (t *Task) CheckValideData(task Task) error {

	executor, ok := PluginRegistry[task.PluginType]
	if !ok {
		return fmt.Errorf("'component.Plugin' плагин для типа '%s' не найден", task.PluginType)
	}
	// Проверяем данные
	if err := executor.ValidateYAML(task.Actions); err != nil {
		return fmt.Errorf("ошибка валидации данных: %v", err)
	}
	return nil
}

// PluginRegistry - глобальная карта для хранения зарегистрированных плагинов
var PluginRegistry = make(map[string]v1.Executor)
