package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/regulatory-transparency-monitor/graph-builder/graph/model"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/dataparser"
	"github.com/regulatory-transparency-monitor/graph-builder/pkg/logger"
	"github.com/spf13/viper"
)

// Neo4jRepository is a Neo4j DB repository
type Neo4jRepository struct {
	Connection neo4j.Driver
}

// NewNeo4jConnection creates a new neo4j connection returns a neo4j.Driver object and an error
func NewNeo4jConnection() (neo4j.Driver, error) {
	target := fmt.Sprintf("%s://%s:%d", viper.GetString("NEO4J_PROTO"), viper.GetString("NEO4J_HOST"), viper.GetInt("NEO4J_PORT"))

	driver, err := neo4j.NewDriver(
		target,
		neo4j.BasicAuth(viper.GetString("NEO4J_USER"), viper.GetString("NEO4J_PASS"), ""),
		func(c *neo4j.Config) {
			c.Encrypted = false
		})

	if err != nil {
		logger.Error("Cannot connect to Neo4j Server", err)
		return nil, err
	}

	logger.Info("Connected to Neo4j Server", logger.LogFields{"neo4j_Instance_uri": target})

	return driver, nil
}

func (r *Neo4jRepository) GetLabels() ([]string, error) {
	session, err := r.Connection.Session(neo4j.AccessModeRead)
	if err != nil {
		return nil, fmt.Errorf("error creating Neo4j session: %v", err)
	}
	defer session.Close()

	query := `
		CALL db.labels()
		`
	result, err := session.Run(query, nil)
	if err != nil {
		return []string{}, fmt.Errorf("error getting labels from Neo4j: %v", err)
	}

	var labels []string
	for result.Next() {
		record := result.Record()
		label, ok := record.Get("label")
		if ok {
			labels = append(labels, label.(string))
		}
	}
	logger.Debug("Got labels from Neo4j", logger.LogFields{"labels": labels})
	return labels, nil
}

func (r *Neo4jRepository) SetupUUIDForKnownLabels() error {
	labels := []string{"Metadata", "Project", "Instance", "Volume", "ClusterNode", "Pod", "PhysicalHost", "PersistentVolume", "PersistentVolumeClaim", "PDIndicator", "DataCategory", "Snapshot"}

	for _, label := range labels {
		if err := r.CreateUUIDConstraints(label); err != nil {
			return err
		}
	}
	return nil
}

func (r *Neo4jRepository) CreateUUIDConstraints(label string) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return fmt.Errorf("error creating Neo4j session: %v", err)
	}
	defer session.Close()

	// Create UUID constraints
	query := fmt.Sprintf("CREATE CONSTRAINT FOR (n:%s) REQUIRE n.uuid IS UNIQUE;", label)
	_, err = session.Run(query, nil)
	if err != nil {
		return fmt.Errorf("error creating UUID constraint for label %s: %v", label, err)
	}
	//logger.Debug("Created UUID constraints for labels", logger.LogFields{"labels": label})

	// Install UUID using APOC
	handlerQuery := fmt.Sprintf("CALL apoc.uuid.install('%s', {addToExistingNodes: true})", label)
	_, err = session.Run(handlerQuery, nil)
	if err != nil {
		return fmt.Errorf("error installing UUID for label %s using APOC: %v", label, err)
	}
	//	logger.Debug("Installed UUID for label using APOC", logger.LogFields{"label": label})

	return nil
}

// GetLatestVersion retrieves the latest version from the Metadata node
func (r *Neo4jRepository) GetLatestVersion() (string, error) {
	session, err := r.Connection.Session(neo4j.AccessModeRead)
	if err != nil {
		return "", fmt.Errorf("error creating Neo4j session: %v", err)
	}
	defer session.Close()

	query := `
        MATCH (m:Metadata)
        RETURN m.version AS version
        ORDER BY m.scanTimestamp DESC
        LIMIT 1
    `

	result, err := session.Run(query, nil)
	if err != nil {
		return "", fmt.Errorf("error getting latest version from metadata node: %s, %v", query, err)
	}

	if result.Next() {
		version, ok := result.Record().Get("version")
		if !ok {
			return "", fmt.Errorf("error getting version from result")
		}
		return version.(string), nil
	}

	return "", fmt.Errorf("no metadata nodes found in the database")
}

// CreateMetadataNode creates a new Metadata node with the provided version and timestamp
func (r *Neo4jRepository) CreateMetadataNode(version string, timestamp string) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return fmt.Errorf("error creating Neo4j session: %v", err)
	}
	defer session.Close()
	// Step 1: Get the latest version
	latestVersion, err := r.GetLatestVersion()
	if err != nil && err.Error() != "no metadata nodes found in the database" {
		// Handle error if it is not just "no metadata nodes found"
		return fmt.Errorf("error getting the latest Metadata version: %v", err)
	}

	query := `
		CREATE (m:Metadata {version: $version, scanTimestamp: $timestamp})
	`

	parameters := map[string]interface{}{
		"version":   version,
		"timestamp": timestamp,
	}

	_, err = session.Run(query, parameters)
	if err != nil {
		return fmt.Errorf("error creating Metadata node in Neo4j with query: %s, %v", query, err)
	}
	// Step 3: If there is a previous version, create a relationship with the new version
	if latestVersion != "" {
		queryCreateRelationship := `
			MATCH (mNew:Metadata {version: $newVersion}), (mOld:Metadata {version: $oldVersion})
			CREATE (mOld)-[:NEXT_VERSION]->(mNew)
		`
		parametersRelationship := map[string]interface{}{
			"newVersion": version,
			"oldVersion": latestVersion,
		}
		_, err = session.Run(queryCreateRelationship, parametersRelationship)
		if err != nil {
			return fmt.Errorf("error creating relationship in Neo4j with query: %s, %v", queryCreateRelationship, err)
		}
	}
	return nil
}

// CreateProject always creates a new project node and returns its UUID
func (r *Neo4jRepository) CreateProjectNode(version string, project dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	query := `
    CREATE (p:Project {
		uuid: apoc.create.uuid(), 
		version: $version,
		id: $id,
		name: $name, 
		type: $type, 
		availabilityZone: $availabilityZone, 
		enabled: $enabled, 
		description: $description
	})
	RETURN p.uuid as uuid
    `

	parameters := map[string]interface{}{
		"version":          version,
		"id":               project.ID,
		"name":             project.Name,
		"type":             project.Type,
		"availabilityZone": project.AvailabilityZone,
		"description":      GetMetadataValue(project.Metadata, "Description", ""),
		"enabled":          GetMetadataValue(project.Metadata, "Enabled", false),
	}

	result, err := session.Run(query, parameters)
	if err != nil {
		return "", fmt.Errorf("error creating Project node: %s, %v", query, err)
	}

	if result.Next() {
		rawUUID, ok := result.Record().Get("uuid")
		if ok && rawUUID != nil {
			uuidStr, ok := rawUUID.(string)
			if ok {
				return uuidStr, nil
			}
		}
	}

	return "", nil
}

// CreateInstance always creates a new instance node
func (r *Neo4jRepository) CreateInstanceNode(version string, instance dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	query := `
    CREATE (i:Instance {
		uuid: apoc.create.uuid(), 
		version: $version,
		id: $id,
		name: $name,
		type: $type,
		availabilityZone: $availabilityZone,
		userID: $userID,
		hostID: $hostID,
		tenantID: $tenantID,
		created: $created,
		updated: $updated,
		volumesAttached: $volumesAttached,
		status: $status
	})
	RETURN i.uuid as uuid
	`

	parameters := map[string]interface{}{
		"version":          version,
		"id":               instance.ID,
		"name":             instance.Name,
		"type":             instance.Type,
		"availabilityZone": instance.AvailabilityZone,
		"userID":           GetMetadataValue(instance.Metadata, "UserID", ""),
		"hostID":           GetMetadataValue(instance.Metadata, "HostID", ""),
		"tenantID":         GetMetadataValue(instance.Metadata, "TenantID", ""),
		"created":          GetMetadataValue(instance.Metadata, "Created", ""),
		"updated":          GetMetadataValue(instance.Metadata, "Updated", ""),
		"volumesAttached":  GetMetadataValue(instance.Metadata, "VolumesAttached", ""),
		"status":           GetMetadataValue(instance.Metadata, "Status", ""),
	}

	result, err := session.Run(query, parameters)
	if err != nil {
		return "", fmt.Errorf("error creating instance node: %s, %v", query, err)
	}

	if result.Next() {
		rawUUID, ok := result.Record().Get("uuid")
		if ok && rawUUID != nil {
			uuidStr, ok := rawUUID.(string)
			if ok {
				return uuidStr, nil
			}
		}
	}

	return "", nil
}

func (r *Neo4jRepository) CreatePhysicalHostNode(version string, host dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	query := `
    CREATE (h:PhysicalHost {
		uuid: apoc.create.uuid(), 
		version: $version,
		id: $id,
		name: $name,
		type: $type,
		availabilityZone: $availabilityZone
	})
	RETURN h.uuid as uuid
	`
	parameters := map[string]interface{}{
		"version":          version,
		"id":               host.ID,
		"name":             host.Name,
		"type":             host.Type,
		"availabilityZone": host.AvailabilityZone,
	}

	_, err = session.Run(query, parameters)

	//logger.Debug("Created physical host in Neo4j", logger.LogFields{"host_id": host.ID}, result)
	if err != nil {
		return "", fmt.Errorf("error creating instance node: %s, %v", query, err)
	}

	return "", nil
}

// CreateVolumeNode always creates a new volume node and returns its UUID
func (r *Neo4jRepository) CreateVolumeNode(version string, volume dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	query := `
    CREATE (v:Volume {
		uuid: apoc.create.uuid(),
		version: $version,
		id: $id,
		name: $name,
		type: $type,
		availabilityZone: $availabilityZone,
		status: $status,
		size: $size,
		bootable: $bootable,
		encrypted: $encrypted,
		multiattach: $multiattach,
		device: $device,
		srcSnapshot: $snapshotID
	})
	RETURN v.uuid as uuid
	`

	parameters := map[string]interface{}{
		"version":          version,
		"id":               volume.ID,
		"name":             volume.Name,
		"type":             volume.Type,
		"availabilityZone": volume.AvailabilityZone,
		"status":           volume.Metadata["status"],
		"size":             volume.Metadata["size"],
		"bootable":         volume.Metadata["bootable"],
		"encrypted":        volume.Metadata["encrypted"],
		"multiattach":      volume.Metadata["multiattach"],
		"device":           volume.Metadata["device"],
		"snapshotID":       volume.Metadata["snapshotID"],
	}

	result, err := session.Run(query, parameters)
	if err != nil {
		return "", fmt.Errorf("error creating Volume node: %s, %v", query, err)
	}

	if result.Next() {
		rawUUID, ok := result.Record().Get("uuid")
		if ok && rawUUID != nil {
			uuidStr, ok := rawUUID.(string)
			if ok {
				return uuidStr, nil
			}
		}
	}
	return "", fmt.Errorf("failed to retrieve UUID for volume: %s", volume.ID)
}

// CreateVolumeNode always creates a new volume node and returns its UUID
func (r *Neo4jRepository) CreateSnapshotNode(version string, snapshot dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	query := `
    CREATE (s:Snapshot {
		uuid: apoc.create.uuid(),
		version: $version,
		id: $id,
		name: $name,
		type: $type,
		status: $status,
		size: $size,
		createdAt: $createdAt,
		updatedAt: $updatedAt,
		description: $description,
		userID: $userID,
		groupSnapshotID: $groupSnapshotID

		
	})
	RETURN s.uuid as uuid
	`

	parameters := map[string]interface{}{
		"version":         version,
		"id":              snapshot.ID,
		"name":            snapshot.Name,
		"type":            snapshot.Type,
		"status":          snapshot.Metadata["Status"],
		"size":            snapshot.Metadata["Size"],
		"createdAt":       snapshot.Metadata["CreatedAt"],
		"updatedAt":       snapshot.Metadata["UpdatedAt"],
		"description":     snapshot.Metadata["Description"],
		"userID":          snapshot.Metadata["UserID"],
		"groupSnapshotID": snapshot.Metadata["GroupSnapshotID"],
	}

	result, err := session.Run(query, parameters)
	if err != nil {
		return "", fmt.Errorf("error creating snapshot node: %s, %v", query, err)
	}

	if result.Next() {
		rawUUID, ok := result.Record().Get("uuid")
		if ok && rawUUID != nil {
			uuidStr, ok := rawUUID.(string)
			if ok {
				return uuidStr, nil
			}
		}
	}
	return "", fmt.Errorf("failed to retrieve UUID for snapshot: %s", snapshot.ID)
}

// CreateClusterNode always creates a new clusterNode and returns its UUID
func (r *Neo4jRepository) CreateClusterNode(version string, clusterNode dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	// Define the CREATE query
	query := `
    CREATE (n:ClusterNode {
		uuid: apoc.create.uuid(), 
		version: $version,
		id: $id,
		name: $name,
		type: $type,
		createdAt: $createdAt
	})
	RETURN n.uuid as uuid
    `

	// Define the parameters
	parameters := map[string]interface{}{
		"version":   version,
		"id":        clusterNode.ID,
		"name":      clusterNode.Name,
		"type":      clusterNode.Type,
		"createdAt": clusterNode.Metadata["CreatedAt"], // assuming createdAt exists in the Metadata
	}

	result, err := session.Run(query, parameters)
	if err != nil {
		return "", fmt.Errorf("error creating ClusterNode: %s, %v", query, err)
	}

	if result.Next() {
		rawUUID, ok := result.Record().Get("uuid")
		if ok && rawUUID != nil {
			uuidStr, ok := rawUUID.(string)
			if ok {
				return uuidStr, nil
			}
		}
	}

	return "", nil
}

// CreatePod always creates a new pod and returns its UUID
func (r *Neo4jRepository) CreatePodNode(version string, pod dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	// Define the CREATE query
	query := `
    CREATE (p:Pod {
		uuid: apoc.create.uuid(), 
		version: $version,
		id: $id,
		name: $name,
		type: $type,
		createdAt: $createdAt,
		storage: $storage
	})
	RETURN p.uuid as uuid
    `

	// Define the parameters
	parameters := map[string]interface{}{
		"version":   version,
		"id":        pod.ID,
		"name":      pod.Name,
		"type":      pod.Type,
		"createdAt": pod.Metadata["CreatedAt"], // assuming createdAt exists in the Metadata
		"storage":   pod.Metadata["Volumes"],
	}

	result, err := session.Run(query, parameters)
	if err != nil {
		return "", fmt.Errorf("error creating Pod: %s, %v", query, err)
	}

	if result.Next() {
		rawUUID, ok := result.Record().Get("uuid")
		if ok && rawUUID != nil {
			uuidStr, ok := rawUUID.(string)
			if ok {
				return uuidStr, nil
			}
		}
	}

	return "", nil
}

func (r *Neo4jRepository) CreatePVNode(version string, pv dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	// Define the CREATE query
	query := `
    CREATE (pv:PersistentVolume {
        uuid: apoc.create.uuid(),
        version: $version,
        id: $id,
        name: $name,
        type: $type,
        createdAt: $createdAt
    })
    RETURN pv.uuid as uuid
    `

	// Define the parameters
	parameters := map[string]interface{}{
		"version":   version,
		"id":        pv.ID,
		"name":      pv.Name,
		"type":      pv.Type,
		"createdAt": pv.Metadata["CreatedAt"],
	}

	result, err := session.Run(query, parameters)
	if err != nil {
		return "", fmt.Errorf("error creating PersistentVolume: %s, %v", query, err)
	}

	if result.Next() {
		rawUUID, ok := result.Record().Get("uuid")
		if ok && rawUUID != nil {
			uuidStr, ok := rawUUID.(string)
			if ok {
				return uuidStr, nil
			}
		}
	}

	return "", nil
}

func (r *Neo4jRepository) CreatePVCNode(version string, pvc dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	// Define the CREATE query
	query := `
    CREATE (pvc:PersistentVolumeClaim {
        uuid: apoc.create.uuid(),
        version: $version,
        id: $id,
        name: $name,
        type: $type
    })
    RETURN pvc.uuid as uuid
    `

	// Define the parameters
	parameters := map[string]interface{}{
		"version": version,
		"id":      pvc.ID,
		"name":    pvc.Name,
		"type":    pvc.Type,
	}

	//logger.Debug(logger.LogFields{"PVC TYPE": pvc.Type})
	result, err := session.Run(query, parameters)
	if err != nil {
		return "", fmt.Errorf("error creating PersistentVolumeClaim: %s, %v", query, err)
	}

	if result.Next() {
		rawUUID, ok := result.Record().Get("uuid")
		if ok && rawUUID != nil {
			uuidStr, ok := rawUUID.(string)
			if ok {
				return uuidStr, nil
			}
		}
	}

	return "", nil
}

func (r *Neo4jRepository) CreatePDNode(version string, pd dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	query := `
    CALL apoc.create.node(['PDIndicator'], {
        uuid: apoc.create.uuid(),
        version: $version,
        id: $id,
        name: $name,
        type: $type
    })
    YIELD node AS pd
    WITH pd, $dataDisclosed.dataCategories AS dataCategories
    UNWIND dataCategories AS category
    CREATE (pd)-[:HAS_CATEGORY]->(:DataCategory {
        name: category.name,
        purpose: category.purpose,
        legalBasis: category.legalBasis,
        storage: category.storage
    })
    RETURN pd.uuid as uuid
    `

	parameters := map[string]interface{}{
		"version": version,
		"id":      pd.ID,
		"name":    pd.Name,
		"type":    pd.Type,
	}

	if pdJSON, ok := pd.Metadata["has_pd"].(string); ok {
		var pdData map[string]interface{}
		if err := json.Unmarshal([]byte(pdJSON), &pdData); err != nil {
			return "", fmt.Errorf("error unmarshalling PD data: %v", err)
		}
		parameters["dataDisclosed"] = pdData
	} else {
		return "", fmt.Errorf("PD data is not a valid JSON string: %v", pd.Metadata["has_pd"])
	}

	result, err := session.Run(query, parameters)
	if err != nil {
		return "", fmt.Errorf("error running Cypher query: %s, %v", query, err)
	}

	if result.Next() {
		return result.Record().GetByIndex(0).(string), nil
	}

	return "", fmt.Errorf("no UUID returned by query: %s", query)
}

// LinkVolumeToInstance creates a relationship between a volume and attached Instances
func (r *Neo4jRepository) LinkVolumeToInstance(volumeUUID string, instanceID string) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	query := `
    MATCH (v:Volume {uuid: $volumeUUID}), (i:Instance {id: $instanceID})
    MERGE (v)-[:AttachedTo]->(i)
    `

	parameters := map[string]interface{}{
		"volumeUUID": volumeUUID,
		"instanceID": instanceID,
	}

	_, err = session.Run(query, parameters)
	if err != nil {
		return fmt.Errorf("error creating AttachedTo relationship: %s, %v", query, err)
	}

	return nil
}

// CreateOrUpdateVolume creates or updates a Volume Node
func (r *Neo4jRepository) CreateOrUpdateVolume(volume dataparser.InfrastructureComponent) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	query := `
		MERGE (v:Volume {ID: $id})
		ON CREATE SET 
			v.Name = $name,
			v.Type = $type,
			v.AvailabilityZone = $availabilityZone,
			v.Status = $status,
			v.Size = $size,
			v.Bootable = $bootable,
			v.Encrypted = $encrypted,
			v.Multiattach = $multiattach,
			v.Device = $device
			
		ON MATCH SET 
			v.Name = $name,
			v.Type = $type,
			v.AvailabilityZone = $availabilityZone,
			v.Status = $status,
			v.Size = $size,
			v.Bootable = $bootable,
			v.Encrypted = $encrypted,
			v.Multiattach = $multiattach,
			v.Device = $device
		`

	parameters := map[string]interface{}{
		"id":               volume.ID,
		"name":             volume.Name,
		"type":             volume.Type,
		"availabilityZone": volume.AvailabilityZone,
		"status":           volume.Metadata["status"], // assuming status exists in the metadata
		"size":             volume.Metadata["size"],
		"bootable":         volume.Metadata["bootable"],
		"encrypted":        volume.Metadata["encrypted"],
		"multiattach":      volume.Metadata["multiattach"],
		"device":           volume.Metadata["device"],
	}

	// Handle relationships
	for _, rel := range volume.Relationships {
		if rel.Type == "AttachedTo" { //  relationship type
			relationshipQuery := `
			MATCH (v:Volume {ID: $volumeID}), (i:Instance {ID: $targetID})
			MERGE (v)-[:AttachedTo]->(i)
			`

			relationshipParameters := map[string]interface{}{
				"volumeID": volume.ID,
				"targetID": rel.Target,
			}

			_, err = session.Run(relationshipQuery, relationshipParameters)
			if err != nil {
				return err
			}
		}

	}

	_, err = session.Run(query, parameters)
	if err != nil {
		logger.Error("error creating Volume in Neo4j", err)
	}
	return err
}

// CreateOrUpdateInstance creates or updates a Instance node
func (r *Neo4jRepository) CreateOrUpdateServer(instance dataparser.InfrastructureComponent) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	query := `
    MERGE (s:Instance {ID: $id})
    ON CREATE SET 
        s.Name = $name,
        s.Type = $type,
        s.AvailabilityZone = $availabilityZone,
        s.UserID = $userID,
        s.HostID = $hostID,
        s.TenantID = $tenantID,
        s.Created = $created,
        s.Updated = $updated,
        s.VolumesAttached = $volumesAttached,
        s.Status = $status
    ON MATCH SET 
        s.Name = $name,
        s.Type = $type,
        s.AvailabilityZone = $availabilityZone,
        s.UserID = $userID,
        s.HostID = $hostID,
        s.TenantID = $tenantID,
        s.Created = $created,
        s.Updated = $updated,
        s.VolumesAttached = $volumesAttached,
        s.Status = $status
    `

	parameters := map[string]interface{}{
		"id":               instance.ID,
		"name":             instance.Name,
		"type":             instance.Type,
		"availabilityZone": instance.AvailabilityZone,
		"userID":           GetMetadataValue(instance.Metadata, "UserID", ""),
		"hostID":           GetMetadataValue(instance.Metadata, "HostID", ""),
		"tenantID":         GetMetadataValue(instance.Metadata, "TenantID", ""),
		"created":          GetMetadataValue(instance.Metadata, "Created", ""),
		"updated":          GetMetadataValue(instance.Metadata, "Updated", ""),
		"volumesAttached":  GetMetadataValue(instance.Metadata, "VolumesAttached", ""),
		"status":           GetMetadataValue(instance.Metadata, "Status", ""),
	}

	_, err = session.Run(query, parameters)
	if err != nil {
		return fmt.Errorf("error creating instance in Neo4j", err)
	} else { // if no err create relationship
		//logger.Debug("Created instance in Neo4j", logger.LogFields{"instance_id": instance.ID})

		// Handle relationships
		for _, rel := range instance.Relationships {
			switch rel.Type {

			case "BelongsTo": // specific relationship type
				relationshipQuery := `
            MATCH (i:Instance {ID: $instanceID}), (p:Project {ID: $targetID})
            MERGE (i)-[:BelongsTo]->(p)
            `

				relationshipParameters := map[string]interface{}{
					"instanceID": instance.ID,
					"targetID":   rel.Target,
				}

				_, err = session.Run(relationshipQuery, relationshipParameters)
				if err != nil {
					return err
				}

			case "AttachedTo":
				relationshipQuery := `
				MATCH (i:Instance {ID: $instanceID}), (v:Volume {ID: $volumeID})
				MERGE (i)-[:AttachedTo]->(v)
				`

				relationshipParameters := map[string]interface{}{
					"instanceID": instance.ID,
					"volumeID":   rel.Target,
				}

				_, err = session.Run(relationshipQuery, relationshipParameters)
				if err != nil {
					return fmt.Errorf("error creating relationship between instance and volume: %v", err)
				}
			}

		}

	}
	return nil
}

func (r *Neo4jRepository) CreateOrUpdateClusterNode(clusterNode dataparser.InfrastructureComponent) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	// Define the MERGE query
	query := `
    MERGE (n:ClusterNode {ID: $id})
    ON CREATE SET 
        n.Name = $name,
        n.Type = $type,
        n.CreatedAt = $createdAt
    ON MATCH SET 
        n.Name = $name,
        n.Type = $type,
        n.CreatedAt = $createdAt
    `

	// Define the parameters
	parameters := map[string]interface{}{
		"id":        clusterNode.ID,
		"name":      clusterNode.Name,
		"type":      clusterNode.Type,
		"createdAt": clusterNode.Metadata["CreatedAt"], // assuming createdAt exists in the Metadata
	}

	_, err = session.Run(query, parameters)
	return err
}

// CreateOrUpdatePod creates or updates a Kubernetes Pod
func (r *Neo4jRepository) CreateOrUpdatePod(pod dataparser.InfrastructureComponent) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	// Define the MERGE query
	query := `
    MERGE (p:Pod {ID: $id})
    ON CREATE SET 
        p.Name = $name,
        p.Type = $type,
        p.CreatedAt = $createdAt
    ON MATCH SET 
        p.Name = $name,
        p.Type = $type,
        p.CreatedAt = $createdAt
    `

	// Define the parameters
	parameters := map[string]interface{}{
		"id":        pod.ID,
		"name":      pod.Name,
		"type":      pod.Type,
		"createdAt": pod.Metadata["CreatedAt"], // assuming createdAt exists in the Metadata
	}

	_, err = session.Run(query, parameters)
	if err != nil {
		logger.Error("Error creating Pod in Neo4j", err)
	} else {
		// Handle relationships if no error while creating pod node
		for _, rel := range pod.Relationships {
			if rel.Type == "RunsOn" {
				relationshipQuery := `
            MATCH (p:Pod {ID: $podID}), (i:Instance {Name: $instanceName})
            MERGE (p)-[:RunsOn]->(i)
            `

				relationshipParameters := map[string]interface{}{
					"podID":        pod.ID,
					"instanceName": rel.Target,
				}

				_, err = session.Run(relationshipQuery, relationshipParameters)
				if err != nil {
					return err
				}
			}
		}
		// logger.Debug("Created pod in Neo4j", logger.LogFields{"pod_id": pod.ID})
	}
	return err
}

// FindInstanceByUUID finds a Instance by its uuid
func (r *Neo4jRepository) GetPdsWithCategory(ctx context.Context, version string, categoryName string) ([]*model.Pod, error) {
	query := `
		MATCH (p:Pod)-[:HAS_PD]->(pd:PDIndicator)-[:HAS_CATEGORY]->(dc:DataCategory)
		WHERE p.version = $version AND dc.name = $categoryName
		RETURN p.id, p.name, p.type, p.createdAt, p.storage
	`
	parameters := map[string]interface{}{
		"version":      version,
		"categoryName": categoryName,
	}

	session, err := r.Connection.Session(neo4j.AccessModeRead)

	if err != nil {
		return nil, err
	}

	defer session.Close()

	result, err := session.Run(query, parameters)
	if err != nil {
		return nil, err
	}
	// logger.Debug("tried to get ndoe", logger.LogFields{"query": query}, logger.LogFields{"parameters": parameters})

	var pods []*model.Pod
	for result.Next() {
		record := result.Record()
		pod := &model.Pod{}
		err := ParseCypherQueryResult(record, "p", pod)
		if err != nil {
			return nil, err // or log error and continue, depending on your use case
		}
		pods = append(pods, pod)
	}

	return pods, nil
}

func (r *Neo4jRepository) FindInstanceByProjectID(ctx context.Context, projectID string) ([]*model.Instance, error) {

	query := `
		match (m:Instance) where m.projectID = $projectID return m.id, m.name, m.created, m.status `

	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	args := map[string]interface{}{
		"projectID": projectID,
	}

	result, err := session.Run(query, args)
	if err != nil {
		logger.Error("Cannot find instances by projectID", logger.LogFields{"projectID": projectID}, err)
		return nil, err
	}

	instances := make([]*model.Instance, 0)

	for result.Next() {
		instance := model.Instance{}
		ParseCypherQueryResult(result.Record(), "m", &instance)
		instances = append(instances, &instance)
	}

	return instances, nil
}

/*// FindInstanceParticipationsByPersonUUID finds people that participated in a Instance
func (r *Neo4jRepository) FindInstanceParticipationsByPersonUUID(ctx context.Context, uuid string) ([]*model.Participation, error) {
	query := `
		match (m:Instance)-[relatedTo]-(p:Person) where p.uuid = $uuid return m.uuid, m.title, m.released, m.tagline, type(relatedTo) as role
	`
	session, err := r.Connection.Session(neo4j.AccessModeWrite)

	if err != nil {
		return nil, err
	}

	defer session.Close()

	args := map[string]interface{}{
		"uuid": uuid,
	}

	result, err := session.Run(query, args)
	if err != nil {
		logger.Error("Cannot find Instances", err)
	}

	logger.Debug("CYPHER_QUERY", logger.LogFields{"query": query, "args": args})

	var participations []*model.Participation

	for result.Next() {
		Instance := models.Instance{}
		ParseCypherQueryResult(result.Record(), "m", &Instance)
		participation := model.Participation{
			Instance: &Instance,
		}
		// Append Role
		if role, ok := result.Record().Get("role"); ok {
			participation.Role = role.(string)
		}

		participations = append(participations, &participation)
	}

	return participations, err
}

// FindPersonByInstanceUUID finds people (actors, directors, writers) by Instance uuid
func (r *Neo4jRepository) FindPersonByInstanceUUID(ctx context.Context, role string, uuid string) ([]*models.Person, error) {
	query := `
		match (p:Person)-[:%s]->(m:Instance)  where m.uuid = $uuid return p.uuid, p.name, p.born
	`
	query = fmt.Sprintf(query, role)

	session, err := r.Connection.Session(neo4j.AccessModeWrite)

	if err != nil {
		return nil, err
	}

	defer session.Close()

	args := map[string]interface{}{
		"uuid": uuid,
		"role": role,
	}

	result, err := session.Run(query, args)
	if err != nil {
		logger.Error("Cannot find any person with that role", err, logger.LogFields{"role": role})
	}

	logger.Debug("CYPHER_QUERY", logger.LogFields{"query": query, "args": args})

	var people []*models.Person

	for result.Next() {
		person := models.Person{}
		ParseCypherQueryResult(result.Record(), "p", &person)
		// Append Role
		person.Role = StringPtr(role)

		people = append(people, &person)
	}

	return people, nil
}
*/
