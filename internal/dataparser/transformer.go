package dataparser

// Transformer interface for transforming raw data into generic data.
type Transformer interface {
	Transform(rawData interface{}) ([]InfrastructureComponent, error)
}


