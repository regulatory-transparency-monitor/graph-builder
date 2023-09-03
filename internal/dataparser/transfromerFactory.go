package dataparser

import (
	"fmt"
	"strings"

	shared "github.com/regulatory-transparency-monitor/commons/models"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
	m "github.com/regulatory-transparency-monitor/kubernetes-provider-plugin/pkg/models"
	"github.com/regulatory-transparency-monitor/openstack-provider-plugin/pkg/models"
)

var TransformerRegistry = make(map[string]Transformer)

func init() {
	TransformerRegistry["os"] = &OpenStackTransformer{}
	TransformerRegistry["k8s"] = &KubernetesTransformer{}
}

type OpenStackTransformer struct{}
type KubernetesTransformer struct{}
type DefaultTransformerFactory struct{}

func TransformData(rawData shared.RawData) ([]InfrastructureComponent, error) {
	var components []InfrastructureComponent

	for key, dataList := range rawData {
		prefix := getPrefix(key) // e.g., "os" from "os_instance" / "aws" from "aws_instance"
		transformer := TransformerRegistry[prefix]

		if transformer == nil {
			logger.Warning(logger.LogFields{"error": "No transformer found for key:", "key": key})
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

func (o *OpenStackTransformer) Transform(key string, data []interface{}) ([]InfrastructureComponent, error) {
	switch key {
	case "os_project":
		return handleProject(data), nil
	case "os_instace":
		return handleCompute(data), nil
	case "os_volume":
		return handleVolume(data), nil
	// ... other cases
	default:
		return nil, fmt.Errorf("unknown key for OpenStack: %s", key)
	}
}

// Transformer for Kubernetes
func (k *KubernetesTransformer) Transform(key string, data []interface{}) ([]InfrastructureComponent, error) {
	switch key {
	case "k8s_node":
		return handleNode(data), nil
	case "k8s_pod":
		return hanldePod(data), nil
	// ... other cases
	default:
		return nil, fmt.Errorf("unknown key for OpenStack: %s", key)
	}

}

func handleProject(data []interface{}) []InfrastructureComponent {
	var components []InfrastructureComponent
	for _, data := range data {
		project, ok := data.(*models.ProjectDetails)
		if !ok {
			fmt.Printf("Expected type models.ProjectDetails, but got: %T\n", data)
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

func handleCompute(data []interface{}) []InfrastructureComponent {
	var components []InfrastructureComponent
	for _, data := range data {

		serverDetails, ok := data.(models.ServerDetails)
		if !ok {
			fmt.Printf("Expected type models.ServerDetails, but got: %T\n", data)
			continue
		}
		server := serverDetails.Server // Accessing the Server field of the ServerDetails struct
		volumeIDs := extractVolumeIDs(server.VolumesAttached)

		var relationships []Relationship
		relationships = append(relationships, Relationship{
			Type:   "BelongsTo",
			Target: server.TenantID, // point to the project
		})

		for _, volumeID := range volumeIDs {
			relationships = append(relationships, Relationship{
				Type:   "AttachedTo",
				Target: volumeID, // point to the volumes
			})
		}

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
				"VolumesAttached": volumeIDs,
			},
			Relationships: relationships,
		}
		components = append(components, component)
	}
	return components
}

// Helper function to extract volume IDs
func extractVolumeIDs(volumesAttached []interface{}) []string {
	var ids []string
	for _, volume := range volumesAttached {
		if vMap, ok := volume.(map[string]interface{}); ok {
			if id, ok := vMap["id"].(string); ok {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func handleVolume(data []interface{}) []InfrastructureComponent {
	var components []InfrastructureComponent

	for _, data := range data {
		volume, ok := data.(*models.Volume)
		if !ok {
			fmt.Printf("Expected type models.Volume, but got: %T\n", data)
			continue
		}
		component := transformVolumeToComponent(volume)

		if component != nil {
			components = append(components, *component)
		}
	}
	return components
}

func transformVolumeToComponent(volume *models.Volume) *InfrastructureComponent {
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

func handleNode(data []interface{}) []InfrastructureComponent {
	var components []InfrastructureComponent
	for _, data := range data {
		nodeList, ok := data.(*m.NodeList)
		if !ok {
			fmt.Printf("Expected type m.NodeList, but got: %T\n", data)
			continue
		}
		nodes := nodeList.Items // Accessing the Server field of the ServerDetails struct
		for _, node := range nodes {
			component := InfrastructureComponent{
				ID:   node.Metadata.UID,
				Name: node.Metadata.Name,
				Type: "ClusterNode",
				Metadata: map[string]interface{}{
					"CreatedAt": node.Metadata.CreationTimestamp,
				},
			}
			components = append(components, component)
		}
	}
	return components
}

func hanldePod(data []interface{}) []InfrastructureComponent {
	var components []InfrastructureComponent
	for _, data := range data {
		podList, ok := data.(*m.PodList)
		if !ok {
			fmt.Printf("Expected type m.PodList, but got: %T\n", data)
			continue
		}
		pods := podList.Items // Accessing the Server field of the ServerDetails struct
		for _, pod := range pods {
			//fmt.Printf("node: %v\n", node)

			component := InfrastructureComponent{
				ID:   pod.Metadata.UID,
				Name: pod.Metadata.Name,
				Type: "Pod",
				Metadata: map[string]interface{}{
					"CreatedAt": pod.Metadata.CreationTime,
				},
				Relationships: []Relationship{
					{
						Type:   "RunsOn",
						Target: pod.Spec.NodeName,
					},
				},
			}
			components = append(components, component)
		}
	}
	return components
}
