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

func (c *Component) CheckValideData(component Component, pc *plugin.PluginController, logMessage func(string, string, ...interface{})) error {

	logMessage("DEBUG", fmt.Sprintf("[Component > %s]>[Valid] Start validation", component.Name))

	if component.Version == "" {
		return fmt.Errorf("[Component > %s]>[Valid] 'version' is empty", component.Name)
	}

	if component.Plugin == "" {
		return fmt.Errorf("[Component > %s]>[Valid] 'plugin' is empty", component.Name)
	}
	if component.ComponentConfig == nil {
		return fmt.Errorf("[Component > %s]>[Valid] 'config' is empty", component.Name)
	}

	logMessage("DEBUG", fmt.Sprintf("[Component > %s]>[Valid] Check executor in ExecutorPluginRegistry...", component.Name))
	executor, ok := pc.ExecutorPluginRegistry[component.Plugin]
	if !ok {
		return fmt.Errorf("'component.Plugin' плагин для типа '%s' не найден", component.Plugin)
	}
	info, _ := executor.GetInfo()
	logMessage("DEBUG", fmt.Sprintf("[Component > %s]>[Valid] Get component for Plugin: %s, componentConfig: %s", component.Name, info, component.ComponentConfig))
	if pluginComponent, err := executor.GetComponent(component.ComponentConfig); err == nil {
		logMessage("DEBUG", fmt.Sprintf("[Component > %s]>[Valid] Validate component for Plugin: %s, pluginComponent: %s", component.Name, info.Name, pluginComponent))
		componentErr := executor.ValidateYAMLComponent(pluginComponent)
		logMessage("DEBUG", fmt.Sprintf("[Component > %s]>[Valid] component %s validate with Plugin: %s", component.Name, info.Name, component.Name))
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

func (c *Common) CheckValideData(common Common, logMessage func(string, string, ...interface{})) error {
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

func (s *Stand) CheckValideData(stand Stand, pc *plugin.PluginController, logMessage func(string, string, ...interface{})) error {

	if stand.Name == "" {
		return fmt.Errorf("[Stand > %s]>[Valid] 'name' is empty", stand.Name)
	}
	if len(stand.Component) == 0 {
		return fmt.Errorf("[Stand > %s]>[Valid] 'components' is empty", stand.Name)
	}

	// Проверка уникальности имен компонентов
	nameSet := make(map[string]bool)
	for _, component := range stand.Component {
		// Проверяем, что имя компонента не пустое
		if component.Name == "" {
			return fmt.Errorf("[Stand > %s]>[Valid] 'component.Name' is empty", stand.Name)
		}

		// Проверяем уникальность имени компонента
		if nameSet[component.Name] {
			return fmt.Errorf("[Stand > %s]>[Valid] Duplicate component.Name found: %s", stand.Name, component.Name)
		}
		nameSet[component.Name] = true

		logMessage("DEBUG", fmt.Sprintf("[Stand > %s]>[Valid] Starting validation for component %s...", stand.Name, component.Name))
		// Проверяем остальные данные компонента
		componentErr := component.CheckValideData(component, pc, logMessage)
		if componentErr != nil {
			return componentErr
		}
	}

	commonErr := stand.Common.CheckValideData(stand.Common, logMessage)
	if commonErr != nil {
		return commonErr
	}

	logMessage("DEBUG", fmt.Sprintf("[Stand > %s]>[Valid] Validation finish!", stand.Name))
	return nil
}

// Структура для файла, содержащего информацию о стенде и группе
type StandsFile struct {
	MsVersion string  `yaml:"msVersion"`
	Release   string  `yaml:"release"`
	Stand     []Stand `yaml:"stand"`
}

func (sf *StandsFile) FindComponent(data map[string]interface{}, logMessage func(string, string, ...interface{})) (map[string]interface{}, error) {
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
	for _, stand := range sf.Stand {
		if stand.Name == searchKey {
			return convertToMap(sf.Stand)
		}

		// Поиск стенда по группе, если указан "group" в данных
		if groupKey, ok := data["group"].(string); ok && groupKey != "" {
			if stand.Group == groupKey {
				return convertToMap(sf.Stand)
			}
		}

		// Поиск по компонентам внутри стенда
		for _, component := range stand.Component {
			// Сравнение имени компонента
			if component.Name == searchKey {
				logMessage("DEBUG", "[StandsFile] Return Component: %s", component.ComponentConfig)
				return component.ComponentConfig, nil
			}

			// Сравнение группы компонента, если указано
			if groupKey, ok := data["group"].(string); ok && component.Group == groupKey {
				return component.ComponentConfig, nil
			}
		}
	}

	// Если ничего не найдено, возвращаем ошибку
	return nil, fmt.Errorf("no component, group, or stand found with name: '%s'", searchKey)
}

func (sf *StandsFile) CascadeValidation(standsFile StandsFile, pc *plugin.PluginController, logMessage func(string, string, ...interface{})) error {

	validErr := sf.ValidateSF(standsFile)
	if validErr != nil {
		return validErr
	}

	for _, stand := range standsFile.Stand {
		logMessage("DEBUG", fmt.Sprintf("[StandsFile]>[Valid] Starting validation 'Stand' '%s'", stand.Name))
		standErr := stand.CheckValideData(stand, pc, logMessage)
		if standErr != nil {
			return standErr
		}
	}
	logMessage("INFO", "[StandsFile]>[Valid] Validation finish!")

	return nil
}

func (sf *StandsFile) ValidateSF(standsFile StandsFile) error {

	if standsFile.MsVersion == "" {
		return fmt.Errorf("[StandsFile]>[Valid] 'msVersion' is empty")
	}
	if standsFile.Release == "" {
		return fmt.Errorf("[StandsFile]>[Valid] 'Release' is empty")
	}
	if len(standsFile.Stand) == 0 {
		return fmt.Errorf("[StandsFile]>[Valid] 'stand' is empty")
	}

	return nil
}
