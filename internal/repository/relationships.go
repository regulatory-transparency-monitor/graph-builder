package repository

/*
// Relationship Handlers
var RelationshipHandlers = map[string]func(*Neo4jRepository, string, string) error{
	"AttachedTo": HandleAttachedToRelationship,
	// add other relationship types and their handlers here
}

// HandleAttachedToRelationship creates an AttachedTo relationship between a volume and a server
func HandleAttachedToRelationship(r *Neo4jRepository, volumeUUID string, targetID string) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	query := `
	MATCH (v:Volume {uuid: $volumeUUID}), (s:Server {id: $targetID})
	MERGE (v)-[:AttachedTo]->(s)
	`
	params := map[string]interface{}{
		"volumeUUID": volumeUUID,
		"targetID":   targetID,
	}

	_, err = session.Run(query, params)
	return err
}

// HandleVolumeRelationships processes all relationships for a given volume
func (r *Neo4jRepository) HandleVolumeRelationships(volume dataparser.InfrastructureComponent, volumeUUID string) error {
	for _, rel := range volume.Relationships {
		if handler, exists := RelationshipHandlers[rel.Type]; exists {
			err := handler(r, volumeUUID, rel.Target)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("relationship type %s not supported", rel.Type)
		}
	}
	return nil
} */
