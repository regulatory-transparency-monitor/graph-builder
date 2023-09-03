package dataparser

// Transformer interface for transforming raw data into generic data.
type Transformer interface {
	Transform(key string, data []interface{}) ([]InfrastructureComponent, error)
}
