package services

import (
	"context"
	"fmt"

	"github.com/regulatory-transparency-monitor/graph-builder/graph/model"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/repository"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/dataparser"
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

func (s *Service) CreateInfrastructureComponent(version string, component dataparser.InfrastructureComponent) (uuid string, err error) {
	switch component.Type {
	case "Project":
		return s.repository.CreateProjectNode(version, component)
	case "Instance":
		return s.repository.CreateInstanceNode(version, component)
	case "Volume":
		return s.repository.CreateVolumeNode(version, component)
	case "ClusterNode":
		return s.repository.CreateClusterNode(version, component)
	case "Pod":
		return s.repository.CreatePodNode(version, component)
	case "PhysicalHost":
		return s.repository.CreatePhysicalHostNode(version, component)
	case "PersistentVolume":
		return s.repository.CreatePVNode(version, component)
	case "PersistentVolumeClaim":
		return s.repository.CreatePVCNode(version, component)
	case "PDIndicator":
		return s.repository.CreatePDNode(version, component)
	case "Snapshot":
		return s.repository.CreateSnapshotNode(version, component)
	default:
		return "", fmt.Errorf("unknown component type: %s", component.Type)
	}
}

func (s *Service) CreateRelationships(v string, component dataparser.InfrastructureComponent) error {
	switch component.Type {
	case "Project":
		return nil
	case "Instance":
		return s.repository.CreateInstanceRelationships(component.ID, v, component.Relationships)
	case "ClusterNode":
		return s.repository.CreateClusterNodeRel(component.ID, v, component.Relationships)
	case "Pod":
		return s.repository.CreatePodRel(component.ID, v, component.Relationships)
	case "Volume":
		return s.repository.CreateVolumeRel(component.ID, v, component.Relationships)
	case "PersistentVolumeClaim":
		return s.repository.CreatePVCRel(component.ID, v, component.Relationships)
	case "PhysicalHost":
		return nil
	case "PDIndicator":
		return nil
	case "PersistentVolume":
		return s.repository.CreatePVRel(component.ID, v, component.Relationships)
	case "Snapshot":
		return s.repository.CreateSnapshotRel(component.ID, v, component.Relationships)
	default:
		return fmt.Errorf("unknown component type: %s", component.Type)
	}
}

func (s *Service) CreatePVRel(pvID string, version string, relationships []dataparser.Relationship) error {
	return s.repository.CreatePVRel(pvID, version, relationships)
}

func (s *Service) CreatePVCRel(pvcID string, version string, relationships []dataparser.Relationship) error {
	return s.repository.CreatePVCRel(pvcID, version, relationships)
}

func (s *Service) CreateVolumeRel(volumeID string, version string, relationships []dataparser.Relationship) error {
	return s.repository.CreateVolumeRel(volumeID, version, relationships)
}

func (s *Service) CreatePodRel(podID string, version string, relationships []dataparser.Relationship) error {
	return s.repository.CreatePodRel(podID, version, relationships)
}
func (s *Service) CreateClusterNodeRel(nodeID string, version string, relationships []dataparser.Relationship) error {
	return s.repository.CreateClusterNodeRel(nodeID, version, relationships)
}
func (s *Service) CreateInstanceRelationships(instanceID string, version string, relationships []dataparser.Relationship) error {
	return s.repository.CreateInstanceRelationships(instanceID, version, relationships)
}
func (s *Service) SetupUUIDForKnownLabels() error {
	return s.repository.SetupUUIDForKnownLabels()
}
func (s *Service) GetLatestVersion() (string, error) {
	return s.repository.GetLatestVersion()
}

func (s *Service) CreateNewMetadataVersion(version string, timestamp string) error {
	return s.repository.CreateMetadataNode(version, timestamp)
}

func (s *Service) CreateMetadataNode(version string, timeString string) error {
	return s.repository.CreateMetadataNode(version, timeString)
}

func (s *Service) LinkProjectToMetadata(version string, projectUUID string) error {
	return s.repository.LinkProjectToMetadata(version, projectUUID)
}

// FindInstanceByUUID finds a Instance by its uuid
func (s *Service) GetPdsWithCategory(ctx context.Context, version string, categoryName string) ([]*model.Pod, error) {
	return s.repository.GetPdsWithCategory(ctx, version, categoryName)
}
