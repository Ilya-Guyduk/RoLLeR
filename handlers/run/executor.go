package run

import (
	"context"
	"fmt"

	//"plugin"
	"reflect"

	v1 "github.com/Ilya-Guyduk/RoLLeR/pei/v1"
)

type Script struct {
	Set        *MigrationSet
	Name       string                 `yaml:"name"`
	PluginType string                 `yaml:"plugin"`
	Actions    map[string]interface{} `yaml:"action"`
	Component  string                 `yaml:"component"`
}

func (s *Script) CheckValideData(script Script) error {

	ctx := context.TODO()

	pc := s.Set.PluginController
	executor, ok := pc.ExecutorPluginRegistry[script.PluginType]
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

func (c *Check) CheckValideData(check Check) (*v1.Check, *v1.Component, error) {

	var pluginCheck v1.Check
	var pluginComponent v1.Component

	ctx := context.TODO()

	pc := c.Set.PluginController

	executor, ok := pc.ExecutorPluginRegistry[check.PluginType]
	if !ok {
		return nil, nil, fmt.Errorf("'component.Plugin' плагин для типа '%s' не найден", check.PluginType)
	}

	if pluginCheck, err := executor.GetCheck(check.Actions); err == nil {
		// Проверяем данные
		if err := executor.ValidateYAMLCheck(ctx, pluginCheck); err != nil {
			return nil, nil, fmt.Errorf("ошибка валидации данных: %v", err)
		}
	}

	component, err := c.Set.StandsFile.FindComponent(c.Component)
	if err != nil {
		return nil, nil, err
	}

	if pluginComponent, err := executor.GetComponent(component); err == nil {
		componentErr := executor.ValidateYAMLComponent(pluginComponent)
		if componentErr != nil {
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}

	return &pluginCheck, &pluginComponent, nil
}

func (c *Check) ExecCheck(item Check, stageName string) error {

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

	// Проверяем данные
	pluginCheck, pluginComponent, CheckValidErr := c.CheckValideData(item)
	if CheckValidErr != nil {
		return fmt.Errorf("ошибка валидации данных: %v", CheckValidErr)
	}

	// Выполняем действие
	if _, err := executor.ExecCheck(ctx, *pluginComponent, *pluginCheck); err != nil {
		return err
	}

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

	return nil
}
