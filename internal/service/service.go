package services

import (
	"context"
	"fmt"

	"github.com/regulatory-transparency-monitor/graph-builder/graph/model"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/dataparser"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/repository"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
)

// Service exposes application bussiness logic
type Service struct {
	repository repository.Repository
}

// NewService creates a new service
func NewService(r repository.Repository) *Service {
	return &Service{
		repository: r,
	}
}

func (s *Service) CreateInfrastructureComponent(component dataparser.InfrastructureComponent) (uuid string, err error) {
	switch component.Type {
	case "Project":
		return s.repository.CreateProjectNode(component)
	case "Server":
		return s.repository.CreateServerNode(component)
	case "Volume":
		return s.repository.CreateVolumeNode(component)
	case "ClusterNode":
		return s.repository.CreateClusterNode(component)
	case "Pod":
		return s.repository.CreatePodNode(component)
	default:
		return "", fmt.Errorf("unknown component type: %s", component.Type)
	}
}

func (s *Service) SetupUUIDForKnownLabels() error {
	return s.repository.SetupUUIDForKnownLabels()
}
func (s *Service) GetLatestVersion() (string, error) {
	return s.repository.GetLatestVersion()
}

func (s *Service) CreateNewMetadataVersion(version, timestamp string) error {
	return s.repository.CreateMetadataNode(version, timestamp)
}

func (s *Service) CreateMetadataNode(version string, timeString string) error {
	return s.repository.CreateMetadataNode(version, timeString)
}

// In the services package
func (s *Service) LinkResourceToMetadata(currentVersion string, projectUUID string) error {
	return s.repository.LinkResourceToMetadata(currentVersion, projectUUID)
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
