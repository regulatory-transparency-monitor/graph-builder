package plugin

import (
	"github.com/regulatory-transparency-monitor/commons/models"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
)

// Plugin interface
type Plugin interface {
	// TODO pass url as parameter
	Initialize() error
	// TODO think about returning pointer or value / rename to FetchData
	FetchData() (models.RawData, error)
}

type PluginConstructor func() Plugin

// Registry for plugin constructors.
var PluginConstructorRegistry = make(map[string]PluginConstructor)

// Registry for initialized plugin instances.
var PluginRegistry = make(map[string]Plugin)

// Returns plugin instance by name from registry.
func GetPluginInstance(name string) (Plugin, error) {
	pluginInstance, exists := PluginRegistry[name]
	if !exists {
		logger.Error("Plugin constructor for %s not found in registry", name)
	}
	return pluginInstance, nil
}
