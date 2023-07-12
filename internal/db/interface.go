package db

import (
	"context"

	"github.com/regulatory-transparency-monitor/graph-builder/graph/model"
)

// Repository definition for repository
type Repository interface {
	// Instance
	FindInstanceByUUID(ctx context.Context, uuid string) (*model.Instance, error)
	FindInstanceByProjectID(ctx context.Context, projectID string) ([]*model.Instance, error)
	TestNeo4jConnection(ctx context.Context) (string, error)
}
