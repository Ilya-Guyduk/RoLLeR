package run

import (
	"context"
	"fmt"

	//"plugin"
	"reflect"

	"github.com/Ilya-Guyduk/RoLLeR/handlers/plugin"
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

func (c *Check) CheckValideData(check Check, pc *plugin.PluginController, stands *StandsFile) (*v1.Check, *v1.Component, error) {

	logMessage("DEBUG", fmt.Sprintf("[Check > %s] Check valide...", check.Name))

	var pluginCheck v1.Check
	var pluginComponent v1.Component

	ctx := context.TODO()

	logMessage("DEBUG", fmt.Sprintf("[Check > %s] Check executor for %s", check.Name, check.PluginType))
	executor, ok := pc.ExecutorPluginRegistry[check.PluginType]
	if !ok {
		return nil, nil, fmt.Errorf("'component.Plugin' плагин для типа '%s' не найден", check.PluginType)
	}

	logMessage("DEBUG", fmt.Sprintf("[Check > %s] GetCheck object for %s", check.Name, check.PluginType))
	if pluginCheck, err := executor.GetCheck(check.Actions); err == nil {
		logMessage("DEBUG", fmt.Sprintf("[Check > %s] Validate Check object for %s", check.Name, check.PluginType))
		if err := executor.ValidateYAMLCheck(ctx, pluginCheck); err != nil {
			return nil, nil, fmt.Errorf("ошибка валидации данных: %v", err)
		} else {
			logMessage("INFO", fmt.Sprintf("[Check > %s] Validate Check for '%s' succes!", check.Name, check.PluginType))
		}
	}
	logMessage("DEBUG", fmt.Sprintf("[Check > %s] Find component for %s", check.Name, check.PluginType))
	logMessage("DEBUG", fmt.Sprintf("[Check > %s] Component %s", check.Name, check.Component))
	component, err := stands.FindComponent(check.Component)
	if err != nil {
		return nil, nil, err
	}
	logMessage("DEBUG", fmt.Sprintf("[Check > %s] GetComponent for Check for %s", check.Name, check.PluginType))
	if pluginComponent, err := executor.GetComponent(component); err == nil {
		logMessage("DEBUG", fmt.Sprintf("[Check > %s] Validate Component object for %s", check.Name, check.PluginType))
		componentErr := executor.ValidateYAMLComponent(pluginComponent)
		if componentErr != nil {
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}

	return &pluginCheck, &pluginComponent, nil
}

func (c *Check) ExecCheck(check Check, stageName string, pc *plugin.PluginController, stands *StandsFile) error {

	logMessage("INFO", fmt.Sprintf("[Check > %s] Starting Check...", check.Name))
	ctx := context.Background()

	logMessage("DEBUG", fmt.Sprintf("[Check > %s] Check executor", check.Name))
	executor, ok := pc.ExecutorPluginRegistry[check.PluginType]
	if !ok {
		return fmt.Errorf("'Check.Plugin' плагин для типа '%s' не найден", check.PluginType)
	} else {
		pluginInfo, err := executor.GetInfo()
		if err == nil {
			logMessage("INFO", fmt.Sprintf("[Check > %s] Plugin: '%s'. Version: %s", check.Name, pluginInfo.Name, pluginInfo.Version))
			logMessage("DEBUG", fmt.Sprintf("[Check > %s] Plugin: '%s'. Desc: %s", check.Name, pluginInfo.Name, pluginInfo.Description))
		} else {
			return err
		}
	}

	v1Check, v1Compomemt, err := check.CheckValideData(check, pc, stands)
	if err != nil {
		return err
	} else {
		checkCode, err := executor.ExecCheck(ctx, v1Compomemt, v1Check)
		if err != nil {
			return nil
		} else if !checkCode {
			logMessage("ERROR", "checkCode is False")
		}
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
