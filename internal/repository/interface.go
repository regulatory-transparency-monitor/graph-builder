package repository

import (
	"context"

	"github.com/regulatory-transparency-monitor/graph-builder/graph/model"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/dataparser"
)

// Repository definition for repository
type Repository interface {
	// Create Update Nodes using generic data
	CreateOrUpdateProject(dataparser.InfrastructureComponent) error
	CreateOrUpdateServer(dataparser.InfrastructureComponent) error
	CreateOrUpdateVolume(dataparser.InfrastructureComponent) error
	// GraphQL logic
	FindInstanceByUUID(ctx context.Context, uuid string) (*model.Instance, error)
	FindInstanceByProjectID(ctx context.Context, projectID string) ([]*model.Instance, error)
	TestNeo4jConnection(ctx context.Context) (string, error)
}
