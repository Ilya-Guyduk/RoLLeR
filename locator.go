package main

import (
	"fmt"
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
	Host HostConfig
}

// findLocation принимает любую структуру и ищет в ней поле "Location".
// Если находит, проверяет Host внутри Location и возвращает его.
// Если не находит, возвращает HostConfig с адресом "localhost" и портом 22.
func findLocation(data interface{}) (*HostConfig, error) {
	val := reflect.ValueOf(data)
	// Проверка на указатель и получение исходного значения.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Ищем поле "Location" в структуре.
	locationField := val.FieldByName("Location")
	if locationField.IsValid() && locationField.Kind() == reflect.Struct {
		hostField := locationField.FieldByName("Host")
		// Проверка, что поле "Host" не пустое и является структурой.
		if hostField.IsValid() && hostField.Kind() == reflect.Struct {
			hostConfig := hostField.Interface().(HostConfig)

			// Проверка наличия заполненного Address, Port, User и Password.
			if hostConfig.Address == "" {
				hostConfig.Address = "localhost"
			}
			if hostConfig.Port == 0 {
				hostConfig.Port = 22
			}
			if hostConfig.User == "" {
				logMessage("WARN", "User не указан")
			}
			if hostConfig.Password == "" {
				logMessage("WARN", "Password не указан")
			}

			logMessage("DEBUG", fmt.Sprintf("Location found - Host: %s", hostConfig.Address))
			return &hostConfig, nil
		} else {
			logMessage("ERROR", "Пустое поле - Host")
		}
	}

	// Если Location или Host не найдены, возвращаем localhost с портом 22.
	logMessage("DEBUG", "Location for connection not found. Using localhost")
	return &HostConfig{
		Address:  "localhost",
		Port:     22,
		User:     "",
		Password: "",
	}, nil
}
