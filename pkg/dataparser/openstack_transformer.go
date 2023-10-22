package dataparser

import (
	"fmt"

	"github.com/regulatory-transparency-monitor/openstack-provider-plugin/pkg/models"
)

// Transformer for OpenStack // here if aws move to gernic Transformer
func (o *OpenStackTransformer) Transform(key string, data []interface{}) ([]InfrastructureComponent, error) {
	switch key {
	case "os_project":
		return handleProject(data), nil
	case "os_instance":
		return handleCompute(data), nil
	case "os_volume":
		return handleVolume(data), nil
	case "os_snapshot":
		return handleSnapshots(data), nil
	default:
		return nil, fmt.Errorf("unknown key for OpenStack: %s", key)
	}
}

// Data mapping functionality:
func handleProject(data []interface{}) []InfrastructureComponent {
	var components []InfrastructureComponent
	for _, data := range data {
		project, ok := data.(*models.ProjectDetails)
		if !ok {
			fmt.Printf("expected type models.ProjectDetails, but got: %T\n", data)
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
	seenHosts := make(map[string]bool)
	var components []InfrastructureComponent
	for _, data := range data {

		instanceDetails, ok := data.(models.ServerDetails)
		if !ok {
			fmt.Printf("Expected type models.instanceDetails, but got: %T\n", data)
			continue
		}
		instance := instanceDetails.Server // Accessing the Instance field of the InstanceDetails struct
		volumeIDs := extractVolumeIDs(instance.VolumesAttached)

		var relationships []Relationship
		relationships = append(relationships,
			Relationship{
				Type:   "BELONGS_TO",
				Target: instance.TenantID, // point to the project
			},
			Relationship{
				Type:   "ASSIGNED_HOST",
				Target: instance.HostID,
			})
		hostID := instance.HostID
		if _, exists := seenHosts[hostID]; !exists {
			hostComponent := InfrastructureComponent{
				ID:               hostID,
				Name:             hostID,
				AvailabilityZone: instance.AvailabilityZone,
				Type:             "PhysicalHost",
			}
			components = append(components, hostComponent)
			seenHosts[hostID] = true
		}

		for _, volumeID := range volumeIDs {
			relationships = append(relationships, Relationship{
				Type:   "ATTACHED_TO",
				Target: volumeID, // point to the volumes
			})
		}

		component := InfrastructureComponent{
			ID:               instance.ID,
			Name:             instance.Name,
			Type:             "Instance",
			AvailabilityZone: instance.AvailabilityZone,
			Metadata: map[string]interface{}{
				"Status":          instance.Status,
				"TenantID":        instance.TenantID,
				"UserID":          instance.UserID,
				"HostID":          instance.HostID,
				"Created":         instance.Created,
				"Updated":         instance.Updated,
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
		volume, ok := data.(models.Volume)
		if !ok {
			fmt.Printf("expected type models.Volume, but got: %T\n", data)
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
	var instanceID string
	if len(volume.Attachments) > 0 {
		instanceID = volume.Attachments[0].ServerID

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
	if volume.SnapshotID != "" {
		metadata["snapshotID"] = volume.SnapshotID
	} else {
		metadata["snapshotID"] = false
	}
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
				Type:   "ATTACHED_TO",
				Target: instanceID,
			},
		},
	}
}

func handleSnapshots(data []interface{}) []InfrastructureComponent {
	var components []InfrastructureComponent
	for _, data := range data {
		snapshot, ok := data.(models.Snapshot)
		if !ok {
			fmt.Printf("expected type models.Snapshot, but got: %T\n", data)
			continue
		}
		component := InfrastructureComponent{
			ID:   snapshot.ID,
			Name: snapshot.Name,
			Type: "Snapshot",
			Metadata: map[string]interface{}{
				"Status":          snapshot.Status,
				"CreatedAt":       snapshot.CreatedAt,
				"UpdatedAt":       snapshot.UpdatedAt,
				"Description":     snapshot.Description,
				"Size":            snapshot.Size,
				"UserID":          snapshot.UserID,
				"GroupSnapshotID": snapshot.GroupSnapshotID,
			},
			Relationships: []Relationship{
				{
					Type:   "SNAPSHOT_OF",
					Target: snapshot.VolumeID,
				},
			},
		}
		components = append(components, component)

	}
	// logger.Debug(logger.LogFields{"snapshot components": components})
	return components
}
