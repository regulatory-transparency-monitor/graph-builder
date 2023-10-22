package dataparser

/*
// Define the OpenStack mapping
var openStackMapping = map[string]string{
    "id":         "server.id",
    "name":       "server.name",
    "status":     "server.status",
    "tenant_id":  "server.tenant_id",
    "created_at": "server.created",
}

// Define the AWS mapping (example, assuming the structure)
var awsMapping = map[string]string{
    "id":         "instance.instance_identifier",
    "name":       "instance.name",
    "status":     "instance.state",
}

func extractDataFromResponse(response []interface{}, mapping map[string]string) []map[string]interface{} {
    var results []map[string]interface{}

    for _, resource := range response {
        result := transform(resource.(map[string]interface{}), mapping)
        results = append(results, result)
    }

    return results
} 
*/
