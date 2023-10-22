package dataparser

import (
	"strings"

	shared "github.com/regulatory-transparency-monitor/commons/models"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
)

// Transformer interface for transforming raw data into generic data.
type Transformer interface {
	Transform(key string, data []interface{}) ([]InfrastructureComponent, error)
}

var TransformerRegistry = make(map[string]Transformer)

type OpenStackTransformer struct{}
type KubernetesTransformer struct {
	pvcToPVMap map[string]string
}
type DefaultTransformerFactory struct{}

func init() {
	TransformerRegistry["os"] = &OpenStackTransformer{}
	TransformerRegistry["k8s"] = &KubernetesTransformer{}
}

// Look for key
func TransformData(rawData shared.RawData) ([]InfrastructureComponent, error) {
	var components []InfrastructureComponent

	for key, dataList := range rawData {
		prefix := getPrefix(key) // e.g., "os" from "os_server" / "aws" from "aws_instance"
		transformer := TransformerRegistry[prefix]

		if transformer == nil {
			logger.Warning(logger.LogFields{"error": "no transformer found for key:", "key": key})
			continue
		}

		transformedData, err := transformer.Transform(key, dataList)
		if err != nil {
			logger.Error(logger.LogFields{"error": "transforming data for:", "key": key})
			continue
		}
		components = append(components, transformedData...)
	}

	return components, nil
}
func getPrefix(key string) string {
	parts := strings.Split(key, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
