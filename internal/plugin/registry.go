package plugin

import (
	"github.com/regulatory-transparency-monitor/commons/models"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
	kubernetesServices "github.com/regulatory-transparency-monitor/kubernetes-provider-plugin/pkg/services"
	openstackServices "github.com/regulatory-transparency-monitor/openstack-provider-plugin/pkg/services"
	"github.com/spf13/viper"
)

// Plugin interface
type Plugin interface {
	// TODO pass url as parameter
	Initialize() error
	// TODO think about returning pointer or value / rename to FetchData
	Scan() (models.RawData, error)
}

type PluginConstructor func() Plugin

// Registry for plugin constructors.
var PluginConstructorRegistry = make(map[string]PluginConstructor)

// Registry for initialized plugin instances.
var PluginRegistry = make(map[string]Plugin)

// Register Plugin Constructors.
func InitConstructor() {
	PluginConstructorRegistry["openstack"] = func() Plugin {
		return &openstackServices.OpenStackPlugin{}
	}
	// TODO add kubernetes plugin
	PluginConstructorRegistry["kubernetes"] = func() Plugin {
		return &kubernetesServices.KubernetesPlugin{}
	}
}

// Initialize enabled Plugins.
func RegisterPlugin() error {
	// Retrieve the list of enabled plugins from config
	providers := viper.Get("providers").([]interface{})

	for _, provider := range providers {
		p := provider.(map[string]interface{})
		name := p["name"].(string)

		if p["enabled"].(bool) {
			pluginConstructor, exists := PluginConstructorRegistry[name]
			if !exists {
				logger.Fatal("Plugin constructor for %s not found in registry", name)
				continue
			}
			pluginInstance := pluginConstructor()
			err := pluginInstance.Initialize()

			if err != nil {
				logger.Error("Error initializing plugin %s: %v", name, err)
				continue // decide whether to continue or halt based on your requirements
			}
			PluginRegistry[name] = pluginInstance
		}
	}
	return nil
}

// Returns plugin instance by name from registry.
func GetPluginInstance(name string) (Plugin, error) {
	pluginInstance, exists := PluginRegistry[name]
	if !exists {
		logger.Error("Plugin constructor for %s not found in registry", name)
	}
	return pluginInstance, nil
}
