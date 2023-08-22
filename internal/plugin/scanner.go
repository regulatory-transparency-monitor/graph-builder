package plugin

import (
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
	"github.com/regulatory-transparency-monitor/openstack-provider-plugin/pkg/services"
	"github.com/spf13/viper"
)

// Fetch data from all enabled plugins.
func Scanner(pluginInstance Plugin) services.CombinedResources {
	d, err := pluginInstance.Scan()
	if err != nil {
		logger.Error("Error scanning plugin %v", err)
	}
	//logger.Info("Data: %v", logger.LogFields{"Provider Plugin response": d.Data})
	logger.Debug("Fetching data ... ")
	return *d
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
