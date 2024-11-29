package v1

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
	Execute() error
	// Возвращает описание действия (для логов и отладки)
	GetDescription() string
}

type actionService interface {
	// Проверяет корректность данных действия
	ValidateYAML(data map[string]interface{}) error
	// Выполняет действие
	Execute() error
	// Возвращает описание действия (для логов и отладки)
	GetDescription() string
}

type Executor interface {
	actionService
	checkService
	Plugin
}
