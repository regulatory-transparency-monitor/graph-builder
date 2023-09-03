package plugin

import (
	"fmt"

	kubernetesServices "github.com/regulatory-transparency-monitor/kubernetes-provider-plugin/pkg/services"
	openstackServices "github.com/regulatory-transparency-monitor/openstack-provider-plugin/pkg/services"
	"github.com/spf13/viper"
)

type PluginManager struct {
	RegisteredPlugins map[string]Plugin
	ActivePlugins     map[string]Plugin
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		RegisteredPlugins: make(map[string]Plugin),
		ActivePlugins:     make(map[string]Plugin),
	}
}

func (pm *PluginManager) RegisterPluginConstructors() {
	PluginConstructorRegistry["openstack"] = func() Plugin {
		return &openstackServices.OpenStackPlugin{}
	}
	PluginConstructorRegistry["kubernetes"] = func() Plugin {
		return &kubernetesServices.KubernetesPlugin{}
	}
}

func (pm *PluginManager) InitializePlugins() error {
	// Retrieve list of enabled plugins from config
	providers := viper.Get("providers").([]interface{})

	for _, provider := range providers {
		p := provider.(map[string]interface{})
		name := p["name"].(string)

		if p["enabled"].(bool) {
			pluginConstructor, exists := PluginConstructorRegistry[name]
			if !exists {
				fmt.Errorf("Plugin constructor for %s not found in registry", name)
				continue
			}
			pluginInstance := pluginConstructor()
			err := pluginInstance.Initialize()

			if err != nil {
				fmt.Printf("Error initializing plugin %s: %v", name, err)
				continue
			}
			pm.ActivePlugins[name] = pluginInstance
		}
	}
	return nil

}

func (pm *PluginManager) GetPlugin(name string) (Plugin, error) {
	pluginInstance, exists := pm.ActivePlugins[name]
	if !exists {
		fmt.Errorf("Plugin instance for %s not found", name)
		return nil, fmt.Errorf("Plugin instance for %s not found", name)
	}
	return pluginInstance, nil
}
