package main

import (
	"errors"
	"reflect"
)

type HostConfig struct {
	Address  string
	Port     int
	User     string
	Password string
}

type KubernetesConfig struct {
	Namespace string `yaml:"namespace"`
}

type Location struct {
	Host       HostConfig       `yaml:"host"`
	KubeConfig KubernetesConfig `yaml:"KubeConfig"`
}

// findHostConfig обрабатывает поле HostConfig внутри структуры Location
// и возвращает его, заполняя значения по умолчанию при необходимости.
func findHostConfig(locationField reflect.Value) (*HostConfig, error) {
	hostField := locationField.FieldByName("Host")
	var hostConfig HostConfig

	if hostField.IsValid() && hostField.Kind() == reflect.Struct {
		hostConfig = hostField.Interface().(HostConfig)
		if hostConfig.Address == "" {
			hostConfig.Address = "localhost"
		}
		if hostConfig.Port == 0 {
			hostConfig.Port = 22
		}
	} else {
		logMessage("DEBUG", "HostConfig не найден в Location")
		return nil, nil
	}

	return &hostConfig, nil
}

// findKubernetesConfig обрабатывает поле KubernetesConfig внутри структуры Location
// и возвращает его только если Namespace не пустой.
func findKubernetesConfig(locationField reflect.Value) (*KubernetesConfig, error) {
	kubeConfigField := locationField.FieldByName("KubeConfig")
	var kubeConfig *KubernetesConfig

	if kubeConfigField.IsValid() && kubeConfigField.Kind() == reflect.Struct {
		kubeConfigVal := kubeConfigField.Interface().(KubernetesConfig)
		// Проверяем, что Namespace не пустой
		if kubeConfigVal.Namespace != "" {
			kubeConfig = &kubeConfigVal
		} else {
			logMessage("DEBUG", "KubernetesConfig пустой, игнорируется")
		}
	} else {
		logMessage("DEBUG", "KubernetesConfig не найден в Location")
		return nil, nil
	}

	return kubeConfig, nil
}

// findLocation обрабатывает любую структуру и ищет в ней поле Location с Host и/или KubeConfig.
// Возвращает HostConfig и KubernetesConfig, если они найдены, иначе значения по умолчанию.
func findLocation(data interface{}) (*HostConfig, *KubernetesConfig, error) {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	locationField := val.FieldByName("Location")
	if locationField.IsValid() && locationField.Kind() == reflect.Struct {
		// Ищем HostConfig
		hostConfig, err := findHostConfig(locationField)
		if err != nil {
			return nil, nil, err
		}

		// Ищем KubernetesConfig
		kubeConfig, err := findKubernetesConfig(locationField)
		if err != nil {
			return nil, nil, err
		}

		// Возвращаем найденные конфигурации
		return hostConfig, kubeConfig, nil
	}

	// Если Location не найдено, возвращаем значение по умолчанию
	return &HostConfig{
		Address:  "localhost",
		Port:     22,
		User:     "",
		Password: "",
	}, nil, errors.New("Location для подключения не найден")
}
