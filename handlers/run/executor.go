package run

import (
	"context"
	"fmt"

	//"plugin"

	"github.com/Ilya-Guyduk/RoLLeR/handlers/plugin"
	v1 "github.com/laplasd/roller-epi/v1"
)

type Check struct {
	Name       string                 `yaml:"name"`
	PluginType string                 `yaml:"plugin"`
	Actions    map[string]interface{} `yaml:"action"`
	Component  map[string]interface{} `yaml:"component"`
}

func (c *Check) CascadeValidation(check Check, pc *plugin.PluginController, stands StandsFile, logMessage func(string, string, ...interface{})) (*v1.Check, *v1.Component, error) {

	validErr := c.ValidateCH(check)
	if validErr != nil {
		return nil, nil, validErr
	}

	var pluginCheck v1.Check
	var pluginComponent v1.Component
	var err error

	ctx := context.TODO()

	logMessage("DEBUG", fmt.Sprintf("[Check:'%s'] Check executor for '%s'", check.Name, check.PluginType))
	executor, ok := pc.ExecutorPluginRegistry[check.PluginType]
	if !ok {
		logMessage("ERROR", fmt.Sprintf("[Check:'%s'] 'Check.Plugin' плагин для типа '%s' не найден", check.Name, check.PluginType))
		err := pc.InstallPlugin(check.PluginType)
		if err != nil {
			return nil, nil, nil
		}
	} else {
		info, _ := executor.GetInfo()
		logMessage("DEBUG", fmt.Sprintf("[Check:'%s'] Executor object for '%s': %s", check.Name, check.PluginType, info))
	}

	logMessage("DEBUG", fmt.Sprintf("[Check:'%s'] GetCheck object for %s", check.Name, check.PluginType))
	if pluginCheck, err = executor.GetCheck(check.Actions); err == nil {

		logMessage("DEBUG", fmt.Sprintf("[Check:'%s'] Validate Check object for %s", check.Name, check.PluginType))
		if err = executor.ValidateYAMLCheck(ctx, pluginCheck); err != nil {

			return nil, nil, fmt.Errorf("ошибка валидации данных: %v", err)
		} else {

			logMessage("INFO", fmt.Sprintf("[Check:'%s'] Validate Check for '%s' succes!", check.Name, check.PluginType))
		}
	}

	logMessage("DEBUG", fmt.Sprintf("[Check:'%s'] Find component for %s", check.Name, check.PluginType))
	logMessage("DEBUG", fmt.Sprintf("[Check:'%s'] Component %s", check.Name, check.Component))
	componentConfig, err := stands.FindComponent(check.Component, logMessage)
	if err != nil {

		return nil, nil, err
	}

	logMessage("DEBUG", fmt.Sprintf("[Check:'%s'] GetComponent for Check for %s, componentConfig: %s", check.Name, check.PluginType, componentConfig))
	if pluginComponent, err = executor.GetComponent(componentConfig); err == nil {

		logMessage("DEBUG", fmt.Sprintf("[Check:'%s'] Validate Component object for %s, pluginComponent: %s", check.Name, check.PluginType, pluginComponent))
		componentErr := executor.ValidateYAMLComponent(pluginComponent)
		if componentErr != nil {

			return nil, nil, fmt.Errorf(" [Check:'%s']'executor.ValidateYAMLComponent' ERROR '%s'", check.Name, err)
		}
	} else {
		return nil, nil, fmt.Errorf(" [Check:'%s']'executor.GetComponent' ERROR '%s'", check.Name, err)
	}
	logMessage("DEBUG", fmt.Sprintf("[Check:'%s'] valitation Finish!", check.Name))

	return &pluginCheck, &pluginComponent, nil
}

func (c *Check) ValidateCH(check Check) error {

	if check.PluginType == "" {
		return fmt.Errorf("[Check:'%s'] 'plugin' is empty", check.Name)
	}
	if check.Component == nil {
		return fmt.Errorf("[Check:'%s'] 'component' is empty", check.Name)
	}
	if check.Actions == nil {
		return fmt.Errorf("[Check:'%s'] 'actions' is empty", check.Name)
	}

	return nil
}

func (c *Check) ExecCheck(check Check, stageName string, pc *plugin.PluginController, stands *StandsFile, logMessage func(string, string, ...interface{})) error {

	logMessage("INFO", fmt.Sprintf("[Check > %s] Start ExecCheck", check.Name))
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

	v1Check, v1Compomemt, err := check.CascadeValidation(check, pc, *stands, logMessage)
	if err != nil {
		return err
	} else {
		checkCode, err := executor.ExecCheck(ctx, *v1Compomemt, *v1Check)
		if err == nil {
			if !checkCode {
				logMessage("ERROR", "checkCode is False")
			}
		} else {
			return err
		}
	}
	return nil
}

type Script struct {
	Name       string                 `yaml:"name"`
	PluginType string                 `yaml:"plugin"`
	Actions    map[string]interface{} `yaml:"action"`
	Component  map[string]interface{} `yaml:"component"`
}

func (s *Script) CascadeValidation(script Script, pc *plugin.PluginController, stands StandsFile, logMessage func(string, string, ...interface{})) (*v1.Action, *v1.Component, error) {

	validErr := s.ValidateSC(script)
	if validErr != nil {
		return nil, nil, validErr
	}

	var pluginAction v1.Action
	var pluginComponent v1.Component
	var err error

	ctx := context.TODO()

	logMessage("DEBUG", fmt.Sprintf("[Script:'%s'] Check executor for '%s'", script.Name, script.PluginType))
	executor, ok := pc.ExecutorPluginRegistry[script.PluginType]
	if !ok {
		logMessage("ERROR", fmt.Sprintf("[Script:'%s'] 'Script.Plugin' плагин для типа '%s' не найден. Попытка установки", script.Name, script.PluginType))
		err = pc.InstallPlugin(script.PluginType)
		if err != nil {
			return nil, nil, err
		}
	} else {
		info, _ := executor.GetInfo()
		logMessage("DEBUG", fmt.Sprintf("[Script:'%s'] Executor for '%s': %s", script.Name, script.PluginType, info))
	}

	logMessage("DEBUG", fmt.Sprintf("[Script:'%s'] GetCheck object for %s", script.Name, script.PluginType))
	if pluginAction, err = executor.GetAction(script.Actions); err == nil {

		logMessage("DEBUG", fmt.Sprintf("[Script:'%s'] Validate Script object for %s", script.Name, script.PluginType))
		if err = executor.ValidateYAMLAction(ctx, pluginAction); err != nil {

			return nil, nil, fmt.Errorf("ошибка валидации данных: %v", err)
		} else {

			logMessage("INFO", fmt.Sprintf("[Script:'%s'] Validate Script for '%s' succes!", script.Name, script.PluginType))
		}
	} else {
		return nil, nil, err
	}

	logMessage("DEBUG", fmt.Sprintf("[Script:'%s'] Find component for %s", script.Name, script.PluginType))
	logMessage("DEBUG", fmt.Sprintf("[Script:'%s'] Component %s", script.Name, script.Component))
	componentConfig, err := stands.FindComponent(script.Component, logMessage)
	if err != nil {

		return nil, nil, err
	}

	logMessage("DEBUG", fmt.Sprintf("[Script:'%s'] GetComponent for Check for %s, componentConfig: %s", script.Name, script.PluginType, componentConfig))
	if pluginComponent, err = executor.GetComponent(componentConfig); err == nil {

		logMessage("DEBUG", fmt.Sprintf("[Script:'%s'] Validate Component object for %s, pluginComponent: %s", script.Name, script.PluginType, pluginComponent))
		componentErr := executor.ValidateYAMLComponent(pluginComponent)
		if componentErr != nil {

			return nil, nil, fmt.Errorf(" [Script:'%s']'executor.ValidateYAMLComponent' ERROR '%s'", script.Name, err)
		}
	} else {
		return nil, nil, fmt.Errorf(" [Script:'%s']'executor.GetComponent' ERROR '%s'", script.Name, err)
	}
	logMessage("DEBUG", fmt.Sprintf("[Script:'%s'] valitation Finish!", script.Name))

	return &pluginAction, &pluginComponent, nil
}

func (s *Script) ValidateSC(script Script) error {

	if script.PluginType == "" {
		return fmt.Errorf("[Script:'%s'] 'plugin' is empty", script.Name)
	}
	if script.Component == nil {
		return fmt.Errorf("[Script:'%s'] 'component' is empty", script.Name)
	}
	if script.Actions == nil {
		return fmt.Errorf("[Script:'%s'] 'actions' is empty", script.Name)
	}

	return nil
}

func (s *Script) ExecScript(script Script, stageName string, pc *plugin.PluginController, stands *StandsFile, logMessage func(string, string, ...interface{})) error {

	ctx := context.Background()

	logMessage("DEBUG", fmt.Sprintf("[Script > %s] Check executor", script.Name))
	executor, ok := pc.ExecutorPluginRegistry[script.PluginType]
	if !ok {
		return fmt.Errorf("'Check.Plugin' плагин для типа '%s' не найден", script.PluginType)
	} else {
		pluginInfo, err := executor.GetInfo()
		if err == nil {
			logMessage("INFO", fmt.Sprintf("[Script > %s] Plugin: '%s'. Version: %s", script.Name, pluginInfo.Name, pluginInfo.Version))
			logMessage("DEBUG", fmt.Sprintf("[Script > %s] Plugin: '%s'. Desc: %s", script.Name, pluginInfo.Name, pluginInfo.Description))
		} else {
			return err
		}
	}

	v1Action, v1Compomemt, err := script.CascadeValidation(script, pc, *stands, logMessage)
	if err != nil {
		return err
	} else {
		err := executor.ExecAction(ctx, *v1Compomemt, *v1Action)
		if err != nil {
			return err
		}
	}
	return nil
}

type Task struct {
	Name       string                 `yaml:"name"`
	PluginType string                 `yaml:"plugin"`
	Actions    map[string]interface{} `yaml:"action"`
	Component  map[string]interface{} `yaml:"component"`
}

func (t *Task) CascadeValidation(task Task, pc *plugin.PluginController, stands StandsFile, logMessage func(string, string, ...interface{})) (*v1.Action, *v1.Component, error) {

	validErr := t.ValidateT(task)
	if validErr != nil {
		return nil, nil, validErr
	}

	var pluginAction v1.Action
	var pluginComponent v1.Component
	var err error

	ctx := context.TODO()

	logMessage("DEBUG", fmt.Sprintf("[Task:'%s'] Check executor for '%s'", task.Name, task.PluginType))
	executor, ok := pc.ExecutorPluginRegistry[task.PluginType]
	if !ok {
		logMessage("ERROR", fmt.Sprintf("[Task:'%s'] 'Task.Plugin' плагин для типа '%s' не найден", task.Name, task.PluginType))
		err = pc.InstallPlugin(task.PluginType)
		if err != nil {
			return nil, nil, err
		}
	} else {
		info, _ := executor.GetInfo()
		logMessage("DEBUG", fmt.Sprintf("[Task:'%s'] Executor object for '%s': %s", task.Name, task.PluginType, info))
	}

	logMessage("DEBUG", fmt.Sprintf("[Task:'%s'] GetAction object for %s", task.Name, task.PluginType))
	if pluginAction, err = executor.GetAction(task.Actions); err == nil {

		logMessage("DEBUG", fmt.Sprintf("[Task:'%s'] Validate Action object for %s", task.Name, task.PluginType))
		if err := executor.ValidateYAMLAction(ctx, pluginAction); err != nil {

			return nil, nil, fmt.Errorf("ошибка валидации данных: %v", err)
		} else {

			logMessage("DEBUG", fmt.Sprintf("[Task:'%s'] Validate Action for '%s' succes!", task.Name, task.PluginType))
		}
	} else {
		return nil, nil, err
	}

	logMessage("DEBUG", fmt.Sprintf("[Task:'%s'] Find component for %s", task.Name, task.PluginType))
	logMessage("DEBUG", fmt.Sprintf("[Task:'%s'] Component %s", task.Name, task.Component))
	componentConfig, err := stands.FindComponent(task.Component, logMessage)
	if err != nil {

		return nil, nil, err
	}

	logMessage("DEBUG", fmt.Sprintf("[Task:'%s'] GetComponent for Action for %s, componentConfig: %s", task.Name, task.PluginType, componentConfig))
	if pluginComponent, err = executor.GetComponent(componentConfig); err == nil {

		logMessage("DEBUG", fmt.Sprintf("[Task:'%s'] Validate Component object for %s, pluginComponent: %s", task.Name, task.PluginType, pluginComponent))
		componentErr := executor.ValidateYAMLComponent(pluginComponent)
		if componentErr != nil {

			return nil, nil, fmt.Errorf(" [Task:'%s']'executor.ValidateYAMLComponent' ERROR '%s'", task.Name, err)
		}
	} else {
		return nil, nil, fmt.Errorf(" [Task:'%s']'executor.GetComponent' ERROR '%s'", task.Name, err)
	}
	logMessage("DEBUG", fmt.Sprintf("[Task:'%s'] valitation Finish!", task.Name))

	return &pluginAction, &pluginComponent, nil
}

func (t *Task) ValidateT(task Task) error {

	if task.PluginType == "" {
		return fmt.Errorf("[Task:'%s'] 'plugin' is empty", task.Name)
	}
	if task.Component == nil {
		return fmt.Errorf("[Task:'%s'] 'component' is empty", task.Name)
	}
	if task.Actions == nil {
		return fmt.Errorf("[Task:'%s'] 'actions' is empty", task.Name)
	}

	return nil
}

func (t *Task) ExecTask(task Task, stageName string, pc *plugin.PluginController, stands *StandsFile, logMessage func(string, string, ...interface{})) error {

	ctx := context.Background()

	logMessage("DEBUG", fmt.Sprintf("[Task > %s] Check executor", task.Name))
	executor, ok := pc.ExecutorPluginRegistry[task.PluginType]
	if !ok {
		return fmt.Errorf("'Task.Plugin' плагин для типа '%s' не найден", task.PluginType)
	} else {
		pluginInfo, err := executor.GetInfo()
		if err == nil {
			logMessage("INFO", fmt.Sprintf("[Task > %s] Plugin: '%s'. Version: %s", task.Name, pluginInfo.Name, pluginInfo.Version))
			logMessage("DEBUG", fmt.Sprintf("[Task > %s] Plugin: '%s'. Desc: %s", task.Name, pluginInfo.Name, pluginInfo.Description))
		} else {
			return err
		}
	}

	v1Action, v1Compomemt, err := task.CascadeValidation(task, pc, *stands, logMessage)
	if err != nil {
		return err
	} else {
		err := executor.ExecAction(ctx, *v1Compomemt, *v1Action)
		if err != nil {
			return err
		}
	}
	return nil
}
