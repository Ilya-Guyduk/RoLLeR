package run

import (
	"fmt"
	"reflect"
)

type Component struct {
	Set             *MigrationSet
	Name            string                 `yaml:"name"`
	Version         string                 `yaml:"version"`
	Group           string                 `yaml:"group"` // Имя группы
	Plugin          string                 `yaml:"plugin"`
	ComponentConfig map[string]interface{} `yaml:"config"`
}

func (c *Component) CheckValideData(component Component) error {

	fmt.Printf("[Component] Start valide\n")
	/*
		if component.Name == "" {
			return fmt.Errorf("'component.Name' name is empty")
		}

		if component.Version == "" {
			return fmt.Errorf("'component.Version' name is empty")
		}

		if component.Plugin == "" {
			return fmt.Errorf("'component.Plugin' name is empty")
		}

		if component.Group == "" {
			return fmt.Errorf("'component.Group' Group is empty")
		}
	*/
	pc := c.Set.PluginController

	executor, ok := pc.ExecutorPluginRegistry[component.Plugin]
	if !ok {
		return fmt.Errorf("'component.Plugin' плагин для типа '%s' не найден", component.Plugin)
	}
	if pluginComponent, err := executor.GetComponent(c.ComponentConfig); err == nil {
		componentErr := executor.ValidateYAMLComponent(pluginComponent)
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

func (s *Stand) CheckValideData(stand Stand) error {

	fmt.Printf("[Stand] Start valide\n")

	if stand.Name == "" {
		return fmt.Errorf("'stand.Name' name is empty")
	}

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

		// Проверяем остальные данные компонента
		componentErr := component.CheckValideData(component)
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

func (s *StandsFile) FindComponent(data interface{}) (map[string]interface{}, error) {

	// Проверяем, является ли входное значение строкой
	searchKey, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("invalid input type: expected string")
	}

	// Преобразователь в map[string]interface{}
	convertToMap := func(obj interface{}) (map[string]interface{}, error) {
		result := make(map[string]interface{})
		// Используем рефлексию для преобразования
		objValue := reflect.ValueOf(obj)
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

	// Поиск стенда по группе
	if s.Stand.Group == searchKey {
		return convertToMap(s.Stand)
	}

	// Поиск по компонентам внутри стенда
	for _, component := range s.Stand.Component {
		// Сравнение имени компонента
		if component.Name == searchKey {
			return convertToMap(component)
		}

		// Сравнение группы компонента
		if component.Group == searchKey {
			return convertToMap(component)
		}
	}

	// Если ничего не найдено, возвращаем ошибку
	return nil, fmt.Errorf("no component, group, or stand found with name: '%s'", searchKey)
}

func (s *StandsFile) CheckValideData(standsFile StandsFile) error {

	fmt.Printf("[StandsFile] Start valide\n")
	/*
		if standsFile.msVersion == "" {
			return fmt.Errorf("'standsFile.msVersion' msVersion is empty")
		}

		if standsFile.Release == "" {
			return fmt.Errorf("'standsFile.Release' Release is empty")
		}
	*/
	fmt.Printf("[StandsFile] Starting valide Stand\n")
	standErr := standsFile.Stand.CheckValideData(standsFile.Stand)
	if standErr != nil {
		return standErr
	}

	return nil
}
