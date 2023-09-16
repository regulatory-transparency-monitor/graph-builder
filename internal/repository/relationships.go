package repository

import (
	"fmt"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/dataparser"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
)

func (r *Neo4jRepository) LinkProjectToMetadata(version string, projectUUID string) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	query := `
	MATCH (m:Metadata {version: $version})
	MATCH (p:Project {uuid: $projectUUID})
	MERGE (m)-[r:SCANNED]->(p)
	RETURN m, r, p;
	`

	parameters := map[string]interface{}{
		"version":     version,
		"projectUUID": projectUUID,
	}

	_, err = session.Run(query, parameters)
	if err != nil {
		return fmt.Errorf("error creating SCANNED relationship: %s, %v", query, err)
	}

	logger.Debug("created SCANNED relationship", logger.LogFields{"version": version, "projectUUID": projectUUID})

	return nil
}

func (r *Neo4jRepository) CreateInstanceRelationships(instanceID string, version string, relationships []dataparser.Relationship) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	for _, rel := range relationships {
		var query string
		parameters := map[string]interface{}{
			"instanceID": instanceID,
			"targetID":   rel.Target,
			"version":    version,
		}

		switch rel.Type {

		case "BELONGS_TO":

			query = `
            MATCH (i:Instance {id: $instanceID, version: $version}), (p:Project {id: $targetID, version: $version})
			MERGE (i)-[:BELONGS_TO]->(p)
			RETURN i, p;

            `
		case "ASSIGNED_HOST":
			query = `
		            MATCH (i:Instance {id: $instanceID, version: $version}), (h:PhysicalHost {id: $targetID, version: $version})
		            MERGE (i)-[:ASSIGNED_HOST]->(h)
		            `
		case "ATTACHED_TO":
			logger.Debug("ATTACHED_TO SERVER relationship", logger.LogFields{"instanceID": instanceID, "targetID": rel.Target, "version": version})
			query = `
					MATCH (i:Instance {id: $instanceID, version: $version}), (v:Volume {id: $targetID, version: $version})
					MERGE (i)-[:ATTACHES]->(v)
					`
		default:
			// Unsupported relationship type
			continue

		}

		_, err := session.Run(query, parameters)
		if err != nil {
			return fmt.Errorf("error creating %s relationship: %v", rel.Type, err)
		}

	}

	return nil
}

func (r *Neo4jRepository) CreateClusterNodeRel(nodeID string, version string, relationships []dataparser.Relationship) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	for _, rel := range relationships {
		var query string
		parameters := map[string]interface{}{
			"nodeID":   nodeID,
			"targetID": rel.Target,
			"version":  version,
		}

		switch rel.Type {

		case "IS_HOSTING":

			query = `
            MATCH (n:ClusterNode {id: $nodeID, version: $version}), (i:Instance {id: $targetID,version: $version})
            MERGE (n)-[:IS_HOSTING]->(i)
			RETURN n, i;
            `

		default:
			// Unsupported relationship type
			continue

		}

		_, err := session.Run(query, parameters)
		if err != nil {
			return fmt.Errorf("error creating %s relationship: %v", rel.Type, err)
		}
	}

	return nil
}

func (r *Neo4jRepository) CreatePodRel(podID string, version string, relationships []dataparser.Relationship) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	for _, rel := range relationships {
		var query string
		parameters := map[string]interface{}{
			"podID":   podID,
			"target":  rel.Target,
			"version": version,
		}

		switch rel.Type {

		case "RUNS_ON":

			query = `
            MATCH (p:Pod {id: $podID,version: $version}), (c:ClusterNode {name: $target, version: $version})
            MERGE (p)-[:RUNS_ON]->(c)
			RETURN p, c;
            `

		default:
			// Unsupported relationship type
			continue

		}

		_, err := session.Run(query, parameters)
		if err != nil {
			return fmt.Errorf("error creating %s relationship: %v", rel.Type, err)
		}

	}

	return nil
}

// HandleAttachedToRelationship creates an AttachedTo relationship between volume and  server
func (r *Neo4jRepository) CreateVolumeRel(volumeID string, version string, relationships []dataparser.Relationship) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	for _, rel := range relationships {
		var query string
		parameters := map[string]interface{}{
			"volumeID": volumeID,
			"target":   rel.Target,
			"version":  version,
		}

		switch rel.Type {

		case "ATTACHED_TO":
			query = `
			MATCH (v:Volume {id: $volumeID, version: $version}), (i:Instance {id: $target, version: $version})
			MERGE (v)-[:ATTACHED_TO]->(i)
            `

		default:
			// Unsupported relationship type
			continue

		}
		_, err := session.Run(query, parameters)
		if err != nil {
			return fmt.Errorf("error creating %s relationship: %v", rel.Type, err)
		}
		logger.Debug(query, logger.LogFields{"volumeID": volumeID, "targetID": rel.Target, "version": version})

	}

	return nil
}
