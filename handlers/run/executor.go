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

	ctx := context.TODO()

	executor, ok := PluginRegistry[script.PluginType]
	if !ok {
		return fmt.Errorf("'script.Plugin' плагин для типа '%s' не найден", script.PluginType)
	}

	var v1Action v1.Action

	// Проверяем данные
	if err := executor.ValidateYAMLAction(ctx, v1Action); err != nil {
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

	ctx := context.TODO()

	pc := c.Set.PluginController

	executor, ok := pc.ExecutorPluginRegistry[check.PluginType]
	if !ok {
		return fmt.Errorf("'component.Plugin' плагин для типа '%s' не найден", check.PluginType)
	}

	if v1Action, err := executor.GetAction(check.Actions); err == nil {
		// Проверяем данные
		if err := executor.ValidateYAMLAction(ctx, v1Action); err != nil {
			return fmt.Errorf("ошибка валидации данных: %v", err)
		}
	}

	component, err := c.Set.StandsFile.FindComponent(c.Component)
	if err != nil {
		return err
	}
	componentErr := executor.ValidateYAMLComponent(component)
	if componentErr != nil {
		return err
	}

	return nil
}

func (c *Check) ExecCheck(item interface{}, stageName string) error {

	ctx := context.TODO()

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
	//actionData := val.FieldByName("Actions").Interface().(map[string]interface{})

	var v1Action v1.Action
	var v1Check v1.Check
	var v1Component v1.Component

	// Проверяем данные
	if err := executor.ValidateYAMLAction(ctx, v1Action); err != nil {
		return fmt.Errorf("ошибка валидации данных: %v", err)
	}

	// Проверяем данные
	if err := executor.ValidateYAMLComponent(componentData); err != nil {
		return fmt.Errorf("ошибка валидации данных: %v", err)
	}

	// Выполняем действие
	//ctx := context.TODO() // Контекст можно адаптировать под требования
	if _, err := executor.ExecuteCheck(ctx, v1Component, v1Check); err != nil {
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
	_, err := t.Set.PluginController.FindExecutorPlugin(t.PluginType)
	if err != nil {
		return err
	}

	// Получаем данные локации и действия
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
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
