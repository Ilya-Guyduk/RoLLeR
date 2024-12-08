package v1

import "context"

// PluginInfo предоставляет общую информацию о плагине.
type PluginInfo struct {
	// Уникальное имя плагина
	Name string
	// Версия плагина
	Version string
	// Описание плагина
	Description string
}

type Component struct{}

// Plugin интерфейс для регистрации и информации о плагине.
type Plugin interface {
	// Возвращает информацию о плагине
	GetInfo() (PluginInfo, error)
}

// actionService интерфейс для действий.
type ActionService interface {
	// Проверяет корректность данных действия
	ValidateYAMLAction(ctx context.Context, data map[string]interface{}) error
	// Выполняет действие с использованием контекста
	ExecuteAction(ctx context.Context, component Component, data map[string]interface{}) error
	ExecuteCheck(ctx context.Context, component Component, data map[string]interface{}) (bool, error)
	// Возвращает описание действия (для логов и отладки)
	GetDescription(ctx context.Context, data map[string]interface{}) string
}

type Executor interface {
	ActionService
	Plugin

	ValidateYAMLComponent(data map[string]interface{}) error
}
