package services

import (
	"context"

	"github.com/regulatory-transparency-monitor/graph-builder/graph/model"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/db"
)

// Service exposes application bussiness logic
type Service struct {
	repository db.Repository
}

// NewService creates a new service
func NewService(r db.Repository) Service {
	return Service{
		repository: r,
	}
}

// FindInstanceByUUID finds a Instance by its uuid
func (s *Service) FindInstanceByUUID(ctx context.Context, uuid string) (*model.Instance, error) {
	return s.repository.FindInstanceByUUID(ctx, uuid)
}

// FindInstanceByUUID finds a Instance by its projectID
func (s *Service) FindInstanceByProjectID(ctx context.Context, projectID string) ([]*model.Instance, error) {
	return s.repository.FindInstanceByProjectID(ctx, projectID)
}

// TestNeo4jConnection tests connectivity to neo4j db
func (s *Service) TestNeo4jConnection(ctx context.Context) (string, error) {
	return s.repository.TestNeo4jConnection(ctx)
}
