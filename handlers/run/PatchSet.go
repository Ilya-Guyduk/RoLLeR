package run

import "github.com/Ilya-Guyduk/RoLLeR/handlers/plugin"

type PatchSet struct {
	StandsFile          *StandsFile
	PluginController    *plugin.PluginController
	ActionMap           map[int]ActionMap
	MigrationSetVersion string   `yaml:"msVersion"`
	Atomic              *bool    `yaml:"atomic"` // Флаг атомарности
	YAMLStandFile       string   `yaml:"stands"` // Путь к файлу стендов
	FromRelease         string   `yaml:"from_release"`
	ToRelease           string   `yaml:"to_release"`
	Stages              []Stages `yaml:"stages"` // Список этапов
}
