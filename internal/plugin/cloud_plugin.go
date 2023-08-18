package plugin

import (
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
	"github.com/regulatory-transparency-monitor/openstack-provider-plugin/pkg/services"
	"github.com/spf13/viper"
)

// Plugin interface
type ProviderPlugin interface {
	// TODO pass url as parameter
	Initialize() error
	Scan() ([]interface{}, error)
}

type PluginConstructor func() ProviderPlugin

var PluginRegistry = make(map[string]PluginConstructor)

func InitPlugins() {
	// Retrieve the list of enabled plugins from config
	providers := viper.Get("providers").([]interface{})
	for _, provider := range providers {
		p := provider.(map[string]interface{})
		name := p["name"].(string)

		if p["enabled"].(bool) {
			switch name {
			case "openstack":
				PluginRegistry[name] = func() ProviderPlugin {
					return &services.OpenStackPlugin{}
				}
				// case "kubernetes":
				//     PluginRegistry[name] = func() ProviderPlugin {
				//         return &services.KubernetesPlugin{}
				//     }
			}
		}
	}

}

// TODO add loop to iterate over all plugins
func RegisterPlugin() ProviderPlugin {
	// TODO iterate for each plugin and initialize it
	p := viper.GetString("providers.0.name")
	enabled := viper.GetBool("providers.0.enabled")
	var pluginInstance ProviderPlugin

	if enabled {
		pluginConstructor, exists := PluginRegistry[p]
		if !exists {
			logger.Fatal("Plugin %s not found in registry", p)
		}
		pluginInstance = pluginConstructor()
		/* err := pluginInstance.Initialize()
		if err != nil {
			logger.Fatal("Error initializing plugin %v", err)
		}  */

	}
	return pluginInstance
}

func StartScan(pluginInstance ProviderPlugin) {

	err := pluginInstance.Initialize()
	if err != nil {
		logger.Error("Error initializing plugin %v", err)
	}
	d, err := pluginInstance.Scan()
	if err != nil {
		logger.Error("Error scanning plugin %v", err)
	} 
	logger.Info("Data: %v", d)

}

func MapURL() map[string]string {
	// Initialize an empty map to store the key-value pairs
	apiMap := make(map[string]string)

	// Retrieve the configuration for the first provider
	providerConfig := viper.GetStringMap("providers.0")

	logger.Info("providerConfig: %v", providerConfig)
	// Check if the 'api_access' key exists and is a map
	if apiAccess, ok := providerConfig["api_access"].(map[string]interface{}); ok {
		// Loop through the api_access keys and store them in the map
		for key, value := range apiAccess {
			// Ensure the value is a string before adding to the map
			if strValue, valid := value.(string); valid {
				apiMap[key] = strValue
			}
		}
	}

	logger.Info("apiMap: %v", apiMap)
	return apiMap
}
