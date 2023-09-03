package repository

import (
	"context"

	"github.com/regulatory-transparency-monitor/graph-builder/graph/model"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/dataparser"
)

// Repository definition for repository
type Repository interface {
	// Metadata logic
	GetLabels() ([]string, error) // Get all labels from the database
	SetupUUIDForKnownLabels() error
	CreateUUIDConstraints(labels string) error
	GetLatestVersion() (string, error)
	CreateMetadataNode(version string, timestamp string) error
	LinkResourceToMetadata(version string, projectUUID string) error
	// Create Update Nodes using generic data
	CreateOrUpdateProject(dataparser.InfrastructureComponent) (uuid string, err error)
	CreateOrUpdateServer(dataparser.InfrastructureComponent) error
	CreateOrUpdateVolume(dataparser.InfrastructureComponent) error
	CreateOrUpdateClusterNode(dataparser.InfrastructureComponent) error
	CreateOrUpdatePod(dataparser.InfrastructureComponent) error
	// GraphQL logic
	FindInstanceByUUID(ctx context.Context, uuid string) (*model.Instance, error)
	FindInstanceByProjectID(ctx context.Context, projectID string) ([]*model.Instance, error)
	TestNeo4jConnection(ctx context.Context) (string, error)
}
