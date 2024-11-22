package plugininterface

// PluginInfo предоставляет общую информацию о плагине.
type PluginInfo struct {
	Name        string // Уникальное имя плагина
	Version     string // Версия плагина
	Description string // Описание плагина
}

// Plugin интерфейс для регистрации и информации о плагине.
type Plugin interface {
	GetInfo() PluginInfo                            // Возвращает информацию о плагине
	Initialize(config map[string]interface{}) error // Инициализация плагина с конфигурацией
}

// Location определяет место выполнения действия (хост, namespace, база данных и т.д.).
type Location interface {
	Validate() error        // Проверяет корректность данных локации
	GetDescription() string // Возвращает описание локации (для логов и отладки)
	Connect() error         // Устанавливает соединение
	Disconnect() error      // Закрывает соединение
}

// Action определяет действие, выполняемое плагином (запуск команды, установка пакета и т.д.).
type Action interface {
	Validate() error        // Проверяет корректность данных действия
	Execute() error         // Выполняет действие
	GetDescription() string // Возвращает описание действия (для логов и отладки)
}

// Connector объединяет работу с локациями и действиями.
type Connector interface {
	Plugin                                                        // Наследует общий интерфейс плагина
	CreateLocation(data map[string]interface{}) (Location, error) // Создаёт объект локации
	CreateAction(data map[string]interface{}) (Action, error)     // Создаёт объект действия
	Execute(location Location, action Action) error               // Выполняет действие в указанной локации
}
