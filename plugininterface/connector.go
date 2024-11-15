package plugininterface

// Connector - интерфейс для плагинов
type Connector interface {
	Connect() error
	Execute(action string) error
}
