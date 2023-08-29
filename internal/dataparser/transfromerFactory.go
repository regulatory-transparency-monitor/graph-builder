package dataparser

import (
	"fmt"
	"reflect"

	shared "github.com/regulatory-transparency-monitor/commons/models"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
	"github.com/regulatory-transparency-monitor/openstack-provider-plugin/pkg/models"
)

type TransformerFactory interface {
	GetTransformer(pluginType string) Transformer
}

type DefaultTransformerFactory struct{}

type OpenStackTransformer struct{}

type KubernetesTransformer struct{}

const (
	ServiceSourceIdentity = "identity"
	ServiceSourceCompute  = "compute"
	ServiceSourceVolume   = "volume"
	ServiceSourceWorkload = "workload"
	ServiceSourceCluster  = "cluster"
)

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

// Transformer is an interface that defines the Transform method for openstack resources.
func (o *OpenStackTransformer) Transform(rawData interface{}) ([]InfrastructureComponent, error) {

	// Transformation logic specific to OpenStack.

	typeName := reflect.TypeOf(rawData).String()
	logger.Info("Type of rawData:", typeName)

	// 1. Type Assertion
	openstackData, ok := rawData.(shared.CombinedResources) // Assuming this is the expected type
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

	// 5 & 6. Return
	return components, nil
}

// Transformer for Kubernetes
func (k *KubernetesTransformer) Transform(rawData interface{}) ([]InfrastructureComponent, error) {
	// Transformation logic specific to Kubernetes.

	typeName := reflect.TypeOf(rawData).String()
	logger.Info("Type of Kubernetes rawData:", typeName)

	// 1. Type Assertion
	kubernetesData, ok := rawData.(shared.CombinedResources) // Assuming this is the expected type
	if !ok {
		return nil, fmt.Errorf("unexpected data type for Kubernetes transformation")
	}

	// 2. Initialize Output
	var components []InfrastructureComponent

	// 3. Loop Through Raw Data and Transform
	for _, resource := range kubernetesData.Data {
		component := transformResourceToComponent(resource)
		//	logger.Debug("Transformed: ", logger.LogFields{"resulting in ": component})
		components = append(components, component...)
	}

	// 4. Handle Relationships (if necessary)

	// 5 & 6. Return
	return components, nil

}

// look at the serviceSoruce
func transformResourceToComponent(data shared.ServiceData) []InfrastructureComponent {
	var components []InfrastructureComponent

	switch data.ServiceSource {
	case ServiceSourceIdentity:
		components = handleIdentity(data)
	case ServiceSourceCompute:
		components = handleCompute(data)
	case ServiceSourceVolume:
		components = handleVolume(data)
	case ServiceSourceCluster:
		components = handleCluster(data)
	default:
		// Handle or log unknown ServiceSource
		logger.Info("Unknown ServiceSource: ", data.ServiceSource)
	}

	return components
}



func handleIdentity(resource shared.ServiceData) []InfrastructureComponent {
	var components []InfrastructureComponent
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
			},
		}
		components = append(components, component)
	}
	return components
}

func handleCluster(resource shared.ServiceData) []InfrastructureComponent {
	var components []InfrastructureComponent
	/* for _, data := range resource.Data {
		// TODO kubePlugin models
		cluster, ok := data.(models.Cluster)
		if !ok {
			continue
		}
		component := InfrastructureComponent{
			ID:   cluster.ID,
			Name: cluster.Name,
			Type: "Cluster",
			Metadata: map[string]interface{}{
				"Description": cluster.Description,
				"Enabled":     cluster.Enabled,
			},
		}
		components = append(components, component)
	} */
	return components
}

func handleCompute(resource shared.ServiceData) []InfrastructureComponent {
	var components []InfrastructureComponent
	for _, data := range resource.Data {
		serverDetails, ok := data.(models.ServerDetails)
		if !ok {
			continue
		}
		server := serverDetails.Server // Accessing the Server field of the ServerDetails struct
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
	return components
}

func handleVolume(resource shared.ServiceData) []InfrastructureComponent {
	var components []InfrastructureComponent

	for _, data := range resource.Data {
		volume, ok := data.(models.Volume)
		if !ok {
			fmt.Println("Error converting data to Volume")
			continue
		}
		component := transformVolumeToComponent(volume)

		if component != nil {
			components = append(components, *component)
		}
	}
	return components
}

func transformVolumeToComponent(volume models.Volume) *InfrastructureComponent {
	metadata := make(map[string]interface{})

	// Handle attachments
	var serverID string
	if len(volume.Attachments) > 0 {
		serverID = volume.Attachments[0].ServerID

		// Add other fields from the attachment to the metadata
		attachment := volume.Attachments[0]
		metadata["attached_at"] = attachment.AttachedAt
		metadata["attachment_id"] = attachment.AttachmentID
		metadata["device"] = attachment.Device
		metadata["host_name"] = attachment.HostName
		metadata["attachment_volume_id"] = attachment.VolumeID
	}

	// Add fields from volume to metadata
	metadata["bootable"] = volume.Bootable
	metadata["consistencygroup_id"] = volume.ConsistencyGroupID
	metadata["description"] = volume.Description
	metadata["encrypted"] = volume.Encrypted
	metadata["metadata"] = volume.Metadata
	metadata["multiattach"] = volume.Multiattach
	metadata["replication_status"] = volume.ReplicationStatus
	metadata["size"] = volume.Size
	metadata["snapshot_id"] = volume.SnapshotID
	metadata["source_volid"] = volume.SourceVolid
	metadata["status"] = volume.Status
	metadata["user_id"] = volume.UserID
	metadata["volume_type"] = volume.VolumeType

	return &InfrastructureComponent{
		ID:               volume.ID,
		Name:             volume.Name,
		Type:             "Volume",
		AvailabilityZone: volume.AvailabilityZone,
		Metadata:         metadata,
		Relationships: []Relationship{
			{
				Type:   "AttachedTo",
				Target: serverID,
			},
		},
	}
}
