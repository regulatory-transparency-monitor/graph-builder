package dataparser

// Transformer interface defines the contract for all transformers.
type Transformer interface {
    Transform(rawData interface{}) ([]InfrastructureComponent, error)
}

// TransformerFactory returns a Transformer based on the plugin/provider type.
func TransformerFactory(pluginType string) Transformer {
    switch pluginType {
    case "openstack":
        return &OpenStackTransformer{}
    case "kubernetes":
        return &KubernetesTransformer{}
    // ... other cases
    default:
        return nil
    }
}

type OpenStackTransformer struct {}
func (o *OpenStackTransformer) Transform(rawData interface{}) ([]InfrastructureComponent, error) {
    // Transformation logic specific to OpenStack.
}

type KubernetesTransformer struct {}
func (k *KubernetesTransformer) Transform(rawData interface{}) ([]InfrastructureComponent, error) {
    // Transformation logic specific to Kubernetes.
}

/* usage
rawData := pluginInstance.Scan()
transformer := TransformerFactory(pluginType)
genericData, err := transformer.Transform(rawData) */