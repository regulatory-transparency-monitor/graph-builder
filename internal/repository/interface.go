package repository

import (
	"context"

	"github.com/regulatory-transparency-monitor/graph-builder/graph/model"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/dataparser"
)

// Repository definition for repository
type Repository interface {
	// Metadata logic
	GetLabels() ([]string, error)                              // Get all labels from the database
	SetupUUIDForKnownLabels() error                            // Create UUID constraints for known labels
	CreateUUIDConstraints(labels string) error                 // Create UUID constraints for a given label
	GetLatestVersion() (string, error)                         // Get the latest version of metaNode from the database
	CreateMetadataNode(version string, timestamp string) error // Create a new metadata node using incremented version

	// Create Nodes using generic data
	CreateProjectNode(version string, project dataparser.InfrastructureComponent) (uuid string, err error)     // Create a new project node
	CreateInstanceNode(version string, instance dataparser.InfrastructureComponent) (uuid string, err error)   // Create a new instance node
	CreatePhysicalHostNode(version string, host dataparser.InfrastructureComponent) (uuid string, err error)   // Create a new physical host node
	CreateVolumeNode(version string, volume dataparser.InfrastructureComponent) (uuid string, err error)       // Create a new volume node
	CreateSnapshotNode(version string, snapshot dataparser.InfrastructureComponent) (uuid string, err error)   // Create a new snapshot node
	CreateClusterNode(version string, clusterNode dataparser.InfrastructureComponent) (uuid string, err error) // Create a new cluster node
	CreatePodNode(version string, pod dataparser.InfrastructureComponent) (uuid string, err error)             // Create a new pod node
	CreatePVNode(version string, pv dataparser.InfrastructureComponent) (uuid string, err error)
	CreatePVCNode(version string, pvc dataparser.InfrastructureComponent) (uuid string, err error)

	//Old loghic Create and update nodes using generic data
	CreateOrUpdateServer(dataparser.InfrastructureComponent) error
	CreateOrUpdateVolume(dataparser.InfrastructureComponent) error
	CreateOrUpdateClusterNode(dataparser.InfrastructureComponent) error
	CreateOrUpdatePod(dataparser.InfrastructureComponent) error

	// Create Relationships
	LinkProjectToMetadata(version string, projectUUID string) error // Link a projectUUID of current scan to metadata node
	// Link Meta to next Metanode
	CreateInstanceRelationships(instanceID string, version string, relationships []dataparser.Relationship) error // Create relationships for a given instance
	CreateClusterNodeRel(nodeID string, version string, relationships []dataparser.Relationship) error            // Create relationships for a given cluster node
	CreateVolumeRel(volumeID string, version string, relationships []dataparser.Relationship) error               // Create relationships for a given volume
	CreatePodRel(podID string, version string, relationships []dataparser.Relationship) error                     // Create relationships for a given pod
	CreatePVCRel(pvcID string, version string, relationships []dataparser.Relationship) error
	CreatePVRel(pvID string, version string, relationships []dataparser.Relationship) error
	CreatePDNode(version string, pd dataparser.InfrastructureComponent) (uuid string, err error)
	CreateSnapshotRel(snapshotID string, version string, relationships []dataparser.Relationship) error //Snapshot to Volume
	// GraphQL API
	GetPdsWithCategory(ctx context.Context, version string, categoryName string) ([]*model.Pod, error) // Use Casae 1
	// Use Casae 2
	// Use Casae 3
}
