package plugin

import (
	"github.com/regulatory-transparency-monitor/commons/models"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
	"github.com/spf13/viper"
)

// Fetch data from all enabled plugins.
func Scanner(pm *PluginManager, pluginName string) models.RawData {
	pluginInstance, err := pm.GetPlugin(pluginName)
	if err != nil {
		logger.Error("Error fetching plugin %v", err)
		return nil
	}

	d, err := pluginInstance.FetchData()
	if err != nil {
		logger.Error("Error scanning plugin %v", err)
	}

	//fmt.Printf("Data from plugin %s: %+v\n", pluginName, d)
	return d
}

// TODO add tor end of req list
func MapURL() map[string]string {
	apiMap := make(map[string]string)

	// Retrieve the configuration for the first provider
	providerConfig := viper.GetStringMap("providers.0")

	logger.Info("providerConfig: %v", providerConfig)
	// Check if the 'api_access' key exists and is a map
	if apiAccess, ok := providerConfig["api_access"].(map[string]interface{}); ok {
		// Loop through the api_access keys and store them in the map
		for key, value := range apiAccess {
			if strValue, valid := value.(string); valid {
				apiMap[key] = strValue
			}
		}
	}

	logger.Info("apiMap: %v", apiMap)
	return apiMap
}
