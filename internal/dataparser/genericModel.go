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

type CombinedResources struct {
	Source string
	Data   []ServiceData
}
type ServiceData struct {
	ServiceSource string
	Data          []interface{}
}
