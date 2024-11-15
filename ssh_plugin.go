package main

import (
	"fmt"

	"github.com/Ilya-Guyduk/RoLLeR/plugininterface" // импортируем наш интерфейс
)

type SSHPlugin struct{}

func (p *SSHPlugin) Connect() error {
	fmt.Println("Подключение через SSH плагин")
	return nil
}

func (p *SSHPlugin) Execute(action string) error {
	fmt.Printf("Выполнение действия %s через SSH плагин\n", action)
	return nil
}

// NewConnector возвращает новый объект, который реализует интерфейс Connector
func NewConnector() plugininterface.Connector {
	return &SSHPlugin{}
}
