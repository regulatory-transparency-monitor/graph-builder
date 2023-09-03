package dataparser

import "time"

type InfrastructureComponent struct {
	ID               string
	Name             string
	Type             string
	AvailabilityZone string
	Metadata         map[string]interface{}
	Relationships    []Relationship
}

type Relationship struct {
	Type   string
	Target string
}
type ScanMetadata struct {
	ScanID   string
	ScanDate time.Time
	Version  string // e.g., the version of the scanner or the source system
}
