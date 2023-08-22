package dataparser

import (
	"fmt"

	"github.com/regulatory-transparency-monitor/openstack-provider-plugin/pkg/models"
	"github.com/regulatory-transparency-monitor/openstack-provider-plugin/pkg/services"
)

type TransformerFactory interface {
	GetTransformer(pluginType string) Transformer
}

type DefaultTransformerFactory struct{}

type OpenStackTransformer struct{}

type KubernetesTransformer struct{}

// TransformerFactory returns a Transformer based on the provider type of the plugin.
func (d *DefaultTransformerFactory) GetTransformer(pluginType string) Transformer {
	switch pluginType {
	case "openstack":
		return &OpenStackTransformer{}
	case "kubernetes":
		return &KubernetesTransformer{}
	// ... other cases
	default:
		return nil
	}
}

func (o *OpenStackTransformer) Transform(rawData interface{}) ([]InfrastructureComponent, error) {

	// Transformation logic specific to OpenStack.

	//typeName := reflect.TypeOf(rawData).String()
	//logger.Info("Type of rawData:", typeName)

	// 1. Type Assertion
	openstackData, ok := rawData.(services.CombinedResources) // Assuming this is the expected type
	if !ok {
		return nil, fmt.Errorf("unexpected data type for OpenStack transformation")
	}

	// 2. Initialize Output
	var components []InfrastructureComponent

	// 3. Loop Through Raw Data and Transform
	for _, resource := range openstackData.Data {
		component := transformResourceToComponent(resource)
		//	logger.Debug("Transformed: ", logger.LogFields{"resulting in ": component})
		components = append(components, component...)
	}

	// 4. Handle Relationships (if necessary)
	// This step is more complex and will depend on the specifics of your data and model.

	// 5 & 6. Return
	return components, nil
}

func transformResourceToComponent(resource services.ServiceData) []InfrastructureComponent {
	var components []InfrastructureComponent

	switch resource.ServiceSource {
	case "identity":
		for _, data := range resource.Data {

			project, ok := data.(models.ProjectDetails)
			if !ok {
				continue
			}
			component := InfrastructureComponent{
				ID:   project.Project.ID,
				Name: project.Project.Name,
				Type: "Project",
				Metadata: map[string]interface{}{
					"Description": project.Project.Description,
					"Enabled":     project.Project.Enabled,
					// ... add more fields as needed
				},
			}
			components = append(components, component)
		}

	case "compute":
		for _, data := range resource.Data {
			serverDetails, ok := data.(models.ServerDetails)
			if !ok {
				continue
			}
			server := serverDetails.Server // Accessing the Server field of the ServerDeitals struct
			component := InfrastructureComponent{
				ID:               server.ID,
				Name:             server.Name,
				Type:             "Server",
				AvailabilityZone: server.AvailabilityZone,
				Metadata: map[string]interface{}{
					"Status":          server.Status,
					"TenantID":        server.TenantID,
					"UserID":          server.UserID,
					"HostID":          server.HostID,
					"Created":         server.Created,
					"Updated":         server.Updated,
					"VolumesAttached": server.VolumesAttached,
				},
				Relationships: []Relationship{
					{
						Type:   "BelongsTo",
						Target: server.TenantID, // pointing to the project
					},
				},
			}
			components = append(components, component)
		}
	}

	return components
}

func (k *KubernetesTransformer) Transform(rawData interface{}) ([]InfrastructureComponent, error) {
	// Transformation logic specific to Kubernetes.

	//typeName := reflect.TypeOf(rawData).String()
	//logger.Info("Type of rawData:", typeName)

	return nil, nil

}
