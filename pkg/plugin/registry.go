package plugin

import (
	"github.com/regulatory-transparency-monitor/commons/models"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
)

// Plugin interface
type Plugin interface {
	Initialize(config map[string]interface{}) error

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
