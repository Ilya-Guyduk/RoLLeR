package main

// LoggingConfig описывает параметры логирования
type LoggingConfig struct {
	Level     string `yaml:"level"`
	Formatter string `yaml:"formatter"`
}

// Plugin описывает параметры плагинов
type PluginConfig struct {
	PluginPath     string `yaml:"plugin_path"`
	PluginRepoPath string `yaml:"plugin_repo_path"`
	DefaultRepo    string `yaml:"default_repo"`
}

type Pei struct {
	Version string `yaml:"version"`
}

// Global глобальные настройки
type Global struct {
	Logging LoggingConfig `yaml:"logging"`
	Plugin  PluginConfig  `yaml:"plugin"`
	Pei     Pei           `yaml:"pei"`
}

// rollerConfig структура конфигурации
type RollerConfig struct {
	Global Global `yaml:"global"`
}
