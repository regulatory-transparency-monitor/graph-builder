package dataparser

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
