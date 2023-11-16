package repository

import (
	"fmt"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/dataparser"
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

	//logger.Debug("created SCANNED relationship", logger.LogFields{"version": version, "projectUUID": projectUUID})

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
			//logger.Debug("ATTACHED_TO SERVER relationship", logger.LogFields{"instanceID": instanceID, "targetID": rel.Target, "version": version})
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

		case "PROVISIONED_BY":

			query = `
            MATCH (n:ClusterNode {id: $nodeID, version: $version}), (i:Instance {id: $targetID,version: $version})
            MERGE (n)-[:PROVISIONED_BY]->(i)
			RETURN n, i;
            `

		default:
			// Unsupported relationship type
			continue

		}

		_, err := session.Run(query, parameters)
		//logger.Debug(query, logger.LogFields{"nodeID": nodeID, "targetID": rel.Target, "version": version})
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
		case "USES_PVC":
			query = `
            MATCH (p:Pod {id: $podID,version: $version}), (pvc:PersistentVolumeClaim {name: $target, version: $version})
            MERGE (p)-[:USES_PVC]->(pvc)
			RETURN p, pvc;
            `
		case "HAS_PD":
			query = `
			MATCH (p:Pod {id: $podID,version: $version}), (pd:PDIndicator {id: $target, version: $version})
            MERGE (p)-[:HAS_PD]->(pd)
			RETURN p, pd;
            `

		default:
			// Unsupported relationship type
			continue

		}

		_, err := session.Run(query, parameters)
		//logger.Debug(query, logger.LogFields{"podID": podID, "targetName": rel.Target, "version": version})
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
		//logger.Debug(query, logger.LogFields{"volumeID": volumeID, "targetID": rel.Target, "version": version})

	}

	return nil
}

// HandleAttachedToRelationship creates an AttachedTo relationship between volume and  server
func (r *Neo4jRepository) CreateSnapshotRel(snapshotID string, version string, relationships []dataparser.Relationship) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	for _, rel := range relationships {
		var query string
		parameters := map[string]interface{}{
			"snapshotID": snapshotID,
			"target":     rel.Target,
			"version":    version,
		}

		switch rel.Type {

		case "SNAPSHOT_OF":
			query = `
			MATCH (s:Snapshot {id: $snapshotID, version: $version}), (v:Volume {id: $target, version: $version})
			MERGE (s)-[:SNAPSHOT_OF]->(v)
            `

		default:
			// Unsupported relationship type
			continue

		}
		_, err := session.Run(query, parameters)
		if err != nil {
			return fmt.Errorf("error creating %s relationship: %v", rel.Type, err)
		}
		//logger.Debug(query, logger.LogFields{"volumeID": volumeID, "targetID": rel.Target, "version": version})

	}

	return nil
}

func (r *Neo4jRepository) CreatePVRel(pvID string, version string, relationships []dataparser.Relationship) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	for _, rel := range relationships {
		var query string
		parameters := map[string]interface{}{
			"pvID":     pvID,
			"targetID": rel.Target,
			"version":  version,
		}

		switch rel.Type {

		case "STORED_ON":
			query = `
            MATCH (pv:PersistentVolume {id: $pvID,version: $version}), (v:Volume {id: $targetID, version: $version})
            MERGE (pv)-[:STORED_ON]->(v)
			RETURN pv, v;
            `
		default:
			// Unsupported relationship type
			continue

		}

		_, err := session.Run(query, parameters)
		//logger.Debug(query, logger.LogFields{"pvID": pvID, "target Volume": rel.Target, "version": version})
		if err != nil {
			return fmt.Errorf("error creating %s relationship: %v", rel.Type, err)
		}

	}

	return nil
}

func (r *Neo4jRepository) CreatePVCRel(pvcID string, version string, relationships []dataparser.Relationship) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	for _, rel := range relationships {
		var query string
		parameters := map[string]interface{}{
			"pvcID":   pvcID,
			"target":  rel.Target,
			"version": version,
		}

		switch rel.Type {

		case "BINDS_TO":
			query = `
            MATCH (pvc:PersistentVolumeClaim {id: $pvcID,version: $version}), (pv:PersistentVolume {name: $target, version: $version})
            MERGE (pvc)-[:BINDS_TO]->(pv)
			RETURN pvc, pv;
            `
		default:
			// Unsupported relationship type
			continue

		}

		_, err := session.Run(query, parameters)
		//logger.Debug(query, logger.LogFields{"pvcID": pvcID, "targetName": rel.Target, "version": version})
		if err != nil {
			return fmt.Errorf("error creating %s relationship: %v", rel.Type, err)
		}

	}

	return nil
}
