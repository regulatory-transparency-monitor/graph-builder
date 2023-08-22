package services

import (
	"context"

	"github.com/regulatory-transparency-monitor/graph-builder/graph/model"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/repository"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
)

// Service exposes application bussiness logic
type Service struct {
	repository repository.Repository
}

// NewService creates a new service
func NewService(r repository.Repository) Service {
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
	logger.Info("Check works")
	return s.repository.TestNeo4jConnection(ctx)
}
