package repository

import (
	"context"

	"github.com/regulatory-transparency-monitor/graph-builder/graph/model"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/dataparser"
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
	CreateProjectNode(dataparser.InfrastructureComponent) (uuid string, err error)             // Create a new project node
	CreateServerNode(server dataparser.InfrastructureComponent) (uuid string, err error)       // Create a new server node
	CreateVolumeNode(volume dataparser.InfrastructureComponent) (uuid string, err error)       // Create a new volume node
	CreateClusterNode(clusterNode dataparser.InfrastructureComponent) (uuid string, err error) // Create a new cluster node
	CreatePodNode(pod dataparser.InfrastructureComponent) (uuid string, err error)                 // Create a new pod node
	// Create and update nodes using generic data
	CreateOrUpdateServer(dataparser.InfrastructureComponent) error
	CreateOrUpdateVolume(dataparser.InfrastructureComponent) error
	CreateOrUpdateClusterNode(dataparser.InfrastructureComponent) error
	CreateOrUpdatePod(dataparser.InfrastructureComponent) error

	// Create Relationships
	LinkResourceToMetadata(version string, projectUUID string) error // Link a projectUUID of current scan to metadata node
	LinkServerToProject(serverUUID string, projectID string) error   // Link a server to a projectID

	// GraphQL logic
	FindInstanceByUUID(ctx context.Context, uuid string) (*model.Instance, error)
	FindInstanceByProjectID(ctx context.Context, projectID string) ([]*model.Instance, error)
	TestNeo4jConnection(ctx context.Context) (string, error)
}
