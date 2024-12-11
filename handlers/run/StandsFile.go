package run

import (
	"fmt"
	//"plugin"
	"reflect"

	"github.com/Ilya-Guyduk/RoLLeR/handlers/plugin"
)

type Component struct {
	Set             *MigrationSet
	Name            string                 `yaml:"name"`
	Version         string                 `yaml:"version"`
	Group           string                 `yaml:"group"` // Имя группы
	Plugin          string                 `yaml:"plugin"`
	ComponentConfig map[string]interface{} `yaml:"config"`
}

func (c *Component) CheckValideData(component Component, pc *plugin.PluginController) error {

	logMessage("DEBUG", fmt.Sprintf("[Component > %s] Check valide...", component.Name))

	if component.Version == "" {
		return fmt.Errorf("'component.Version' name is empty")
	}

	if component.Plugin == "" {
		return fmt.Errorf("'component.Plugin' name is empty")
	}

	if component.Group == "" {
		return fmt.Errorf("'component.Group' Group is empty")
	}

	logMessage("DEBUG", fmt.Sprintf("[Component > %s] Check executor in ExecutorPluginRegistry...", component.Name))
	executor, ok := pc.ExecutorPluginRegistry[component.Plugin]
	if !ok {
		return fmt.Errorf("'component.Plugin' плагин для типа '%s' не найден", component.Plugin)
	}
	info, _ := executor.GetInfo()
	logMessage("DEBUG", fmt.Sprintf("[Component > %s] Get component for Plugin: %s", component.Name, info.Name))
	if pluginComponent, err := executor.GetComponent(c.ComponentConfig); err == nil {
		logMessage("DEBUG", fmt.Sprintf("[Component > %s] Validate component for Plugin: %s", component.Name, info.Name))
		componentErr := executor.ValidateYAMLComponent(pluginComponent)
		logMessage("DEBUG", fmt.Sprintf("[Component > %s] component %s validate with Plugin: %s", component.Name, info.Name, component.Name))
		if componentErr != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

type Common struct {
}

func (c *Common) CheckValideData(common Common) error {
	return nil
}

type Stand struct {
	Name        string      `yaml:"name"`   // Имя стенда
	Description string      `yaml:"desc"`   // Краткое описание
	Group       string      `yaml:"group"`  // Имя группы
	Common      Common      `yaml:"common"` // Дополнительные настройки
	Include     []string    `yaml:"include"`
	Component   []Component `yaml:"components"`
}

func (s *Stand) CheckValideData(stand Stand, pc *plugin.PluginController) error {

	logMessage("DEBUG", fmt.Sprintf("[Stand > %s] Check valide stand...", stand.Name))

	if stand.Name == "" {
		return fmt.Errorf("'stand.Name' name is empty")
	}
	logMessage("DEBUG", fmt.Sprintf("[Stand > %s] stand.Name - %s", stand.Name, stand.Name))

	// Проверка уникальности имен компонентов
	nameSet := make(map[string]bool)
	for _, component := range stand.Component {
		// Проверяем, что имя компонента не пустое
		if component.Name == "" {
			return fmt.Errorf("component.Name is empty")
		}

		// Проверяем уникальность имени компонента
		if nameSet[component.Name] {
			return fmt.Errorf("duplicate component.Name found: %s", component.Name)
		}
		nameSet[component.Name] = true

		logMessage("DEBUG", fmt.Sprintf("[Stand > %s] Starting valide for component %s...", stand.Name, component.Name))
		// Проверяем остальные данные компонента
		componentErr := component.CheckValideData(component, pc)
		if componentErr != nil {
			return componentErr
		}
	}

	commonErr := stand.Common.CheckValideData(stand.Common)
	if commonErr != nil {
		return commonErr
	}

	return nil
}

// Структура для файла, содержащего информацию о стенде и группе
type StandsFile struct {
	msVersion string `yaml:"msVersion"`
	Release   string `yaml:"release"`
	Stand     Stand  `yaml:"stand"`
}

func (s *StandsFile) FindComponent(data map[string]interface{}) (map[string]interface{}, error) {
	logMessage("DEBUG", "[StandsFile] Find Component...")

	// Проверяем наличие ключа "name" в данных
	searchKey, ok := data["name"].(string)
	if !ok || searchKey == "" {
		return nil, fmt.Errorf("invalid input: 'name' field is required and must be a string")
	}

	// Преобразователь в map[string]interface{}
	convertToMap := func(obj interface{}) (map[string]interface{}, error) {
		result := make(map[string]interface{})
		// Используем рефлексию для преобразования
		objValue := reflect.ValueOf(obj)
		if objValue.Kind() == reflect.Ptr {
			objValue = objValue.Elem()
		}
		if objValue.Kind() != reflect.Struct {
			return nil, fmt.Errorf("cannot convert non-struct type to map")
		}
		objType := objValue.Type()
		for i := 0; i < objValue.NumField(); i++ {
			fieldName := objType.Field(i).Name
			fieldValue := objValue.Field(i).Interface()
			result[fieldName] = fieldValue
		}
		return result, nil
	}

	// Поиск стенда по имени
	if s.Stand.Name == searchKey {
		return convertToMap(s.Stand)
	}

	// Поиск стенда по группе, если указан "group" в данных
	if groupKey, ok := data["group"].(string); ok && groupKey != "" {
		if s.Stand.Group == groupKey {
			return convertToMap(s.Stand)
		}
	}

	// Поиск по компонентам внутри стенда
	for _, component := range s.Stand.Component {
		// Сравнение имени компонента
		if component.Name == searchKey {
			return convertToMap(component)
		}

		// Сравнение группы компонента, если указано
		if groupKey, ok := data["group"].(string); ok && component.Group == groupKey {
			return convertToMap(component)
		}
	}

	// Если ничего не найдено, возвращаем ошибку
	return nil, fmt.Errorf("no component, group, or stand found with name: '%s'", searchKey)
}

func (s *StandsFile) CheckValideData(standsFile StandsFile, pc *plugin.PluginController) error {

	logMessage("DEBUG", fmt.Sprintf("[StandsFile] Check valide..."))
	/*
		if standsFile.msVersion == "" {
			return fmt.Errorf("'standsFile.msVersion' msVersion is empty")
		}

		if standsFile.Release == "" {
			return fmt.Errorf("'standsFile.Release' Release is empty")
		}
	*/
	logMessage("DEBUG", fmt.Sprintf("[StandsFile] Starting valide Stand"))
	standErr := standsFile.Stand.CheckValideData(standsFile.Stand, pc)
	if standErr != nil {
		return standErr
	}

	return nil
}
