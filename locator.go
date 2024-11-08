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

type Location struct {
	Host HostConfig
}

// findLocation принимает любую структуру и ищет в ней поле "Location".
// Если находит, проверяет Host внутри Location и возвращает его, или ошибку, если поля не заполнены.
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

			// Проверка адреса
			if hostConfig.Address == "" {
				return nil, errors.New("адрес не указан")
			}

			// Проверка и установка порта по умолчанию (22)
			if hostConfig.Port == 0 {
				hostConfig.Port = 22
			}

			// Проверка пользователя
			if hostConfig.User == "" {
				return nil, errors.New("пользователь не указан")
			}

			// Проверка пароля
			if hostConfig.Password == "" {
				return nil, errors.New("пароль не указан")
			}

			return &hostConfig, nil
		} else {
			return nil, errors.New("пустое поле - Host")
		}
	}

	// Если Location или Host не найдены
	return nil, errors.New("Location для подключения не найден")
}
