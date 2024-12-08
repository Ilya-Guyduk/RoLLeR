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
type Action struct{}
type Check struct{}

// Plugin интерфейс для регистрации и информации о плагине.
type Plugin interface {
	// Возвращает информацию о плагине
	GetInfo() (PluginInfo, error)
}

// actionService интерфейс для действий.
type ActionService interface {
	// Проверяет корректность данных действия
	ValidateYAMLAction(ctx context.Context, action Action) error

	GetAction(data map[string]interface{}) (Action, error)
	// Выполняет действие с использованием контекста
	ExecuteAction(ctx context.Context, component Component, action Action) error
	// Выполняет действие с использованием контекста
	ExecuteCheck(ctx context.Context, component Component, check Check) (bool, error)
	// Возвращает описание действия (для логов и отладки)
	GetDescription(ctx context.Context, data map[string]interface{}) string
}

type ComponentService interface {
	ValidateYAMLComponent(data map[string]interface{}) error
	GetComponent(data map[string]interface{}) (Component, error)
}

type Executor interface {
	ComponentService
	ActionService
	Plugin
}
