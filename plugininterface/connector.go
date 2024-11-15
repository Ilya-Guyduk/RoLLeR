package plugininterface

// LocationInterface представляет интерфейс для работы с Location
type LocationInterface interface {
	Validate() error // Проверяет корректность данных локации
}

// ActionInterface представляет интерфейс для работы с Action
type ActionInterface interface {
	Execute() error // Выполняет действие
}

// Connector определяет интерфейс плагина
type Connector interface {
	Connect(location LocationInterface) error                              // Устанавливает соединение
	Execute(action ActionInterface) error                                  // Выполняет действие через плагин
	CreateLocation(data map[string]interface{}) (LocationInterface, error) // Создаёт Location
	CreateAction(data map[string]interface{}) (ActionInterface, error)     // Создаёт Action
}
