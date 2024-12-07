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

// Plugin интерфейс для регистрации и информации о плагине.
type Plugin interface {
	// Возвращает информацию о плагине
	GetInfo() (PluginInfo, error)
}

type checkService interface {
	// Проверяет корректность данных действия
	ValidateYAML(data map[string]interface{}) error
	// Выполняет действие
	Execute(ctx context.Context) error
	// Возвращает описание действия (для логов и отладки)
	GetDescription(data map[string]interface{}) string
}

// Status представляет результат выполнения действия.
type Status struct {
	Success bool   // Флаг успеха
	Message string // Сообщение об ошибке или успехе
}
type componentService interface {
	ValidateYAMLComponent(data map[string]interface{}) error
}

// actionService интерфейс для действий.
type actionService interface {
	// Проверяет корректность данных действия
	ValidateYAML(data map[string]interface{}) error
	// Выполняет действие с использованием контекста
	Execute(ctx context.Context) error
	// Возвращает описание действия (для логов и отладки)
	GetDescription(data map[string]interface{}) string
	// Возвращает статус последнего выполнения действия
	GetStatus() Status
}

type Executor interface {
	componentService
	actionService
	checkService
	Plugin
}
