package repository

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/regulatory-transparency-monitor/graph-builder/graph/model"
	"github.com/regulatory-transparency-monitor/graph-builder/internal/dataparser"
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

	logger.Info("Connected to Neo4j Server", logger.LogFields{"neo4j_server_uri": target})

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
	labels := []string{"Metadata", "Project", "Server", "Volume", "ClusterNode", "Pod"}

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
	logger.Debug("Created UUID constraints for labels", logger.LogFields{"labels": label})

	// Install UUID using APOC
	handlerQuery := fmt.Sprintf("CALL apoc.uuid.install('%s', {addToExistingNodes: true})", label)
	_, err = session.Run(handlerQuery, nil)
	if err != nil {
		return fmt.Errorf("error installing UUID for label %s using APOC: %v", label, err)
	}
	//logger.Debug("Installed UUID for label using APOC", logger.LogFields{"label": label})

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
        RETURN m.Version AS version
        ORDER BY m.ScanTimestamp DESC
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

	query := `
		CREATE (m:Metadata {Version: $version, ScanTimestamp: $timestamp})
	`

	parameters := map[string]interface{}{
		"version":   version,
		"timestamp": timestamp,
	}

	_, err = session.Run(query, parameters)
	if err != nil {
		return fmt.Errorf("error creating Metadata node in Neo4j with query: %s, %v", query, err)
	}
	return nil
}

// CreateProject always creates a new project node and returns its UUID
func (r *Neo4jRepository) CreateProjectNode(project dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	query := `
    CREATE (p:Project {
		uuid: apoc.create.uuid(), 
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

// CreateServer always creates a new server node
func (r *Neo4jRepository) CreateServerNode(server dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	query := `
    CREATE (s:Server {
		uuid: apoc.create.uuid(), 
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
	RETURN s.uuid as uuid
	`

	parameters := map[string]interface{}{
		"id":               server.ID,
		"name":             server.Name,
		"type":             server.Type,
		"availabilityZone": server.AvailabilityZone,
		"userID":           GetMetadataValue(server.Metadata, "UserID", ""),
		"hostID":           GetMetadataValue(server.Metadata, "HostID", ""),
		"tenantID":         GetMetadataValue(server.Metadata, "TenantID", ""),
		"created":          GetMetadataValue(server.Metadata, "Created", ""),
		"updated":          GetMetadataValue(server.Metadata, "Updated", ""),
		"volumesAttached":  GetMetadataValue(server.Metadata, "VolumesAttached", ""),
		"status":           GetMetadataValue(server.Metadata, "Status", ""),
	}
	result, err := session.Run(query, parameters)
	if err != nil {
		return "", fmt.Errorf("error creating server node: %s, %v", query, err)
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

// CreateVolumeNode always creates a new volume node and returns its UUID
func (r *Neo4jRepository) CreateVolumeNode(volume dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	query := `
    CREATE (v:Volume {
		uuid: apoc.create.uuid(),
		id: $id,
		name: $name,
		type: $type,
		availabilityZone: $availabilityZone,
		status: $status,
		size: $size,
		bootable: $bootable,
		encrypted: $encrypted,
		multiattach: $multiattach,
		device: $device
	})
	RETURN v.uuid as uuid
	`

	parameters := map[string]interface{}{
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

// CreateClusterNode always creates a new clusterNode and returns its UUID
func (r *Neo4jRepository) CreateClusterNode(clusterNode dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	// Define the CREATE query
	query := `
    CREATE (n:ClusterNode {
		uuid: apoc.create.uuid(), 
		id: $id,
		name: $name,
		type: $type,
		createdAt: $createdAt
	})
	RETURN n.uuid as uuid
    `

	// Define the parameters
	parameters := map[string]interface{}{
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
func (r *Neo4jRepository) CreatePodNode(pod dataparser.InfrastructureComponent) (uuid string, err error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	// Define the CREATE query
	query := `
    CREATE (p:Pod {
		uuid: apoc.create.uuid(), 
		id: $id,
		name: $name,
		type: $type,
		createdAt: $createdAt
	})
	RETURN p.uuid as uuid
    `

	// Define the parameters
	parameters := map[string]interface{}{
		"id":        pod.ID,
		"name":      pod.Name,
		"type":      pod.Type,
		"createdAt": pod.Metadata["CreatedAt"], // assuming createdAt exists in the Metadata
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

func (r *Neo4jRepository) LinkResourceToMetadata(version string, projectUUID string) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	query := `
	MATCH (m:Metadata {Version: $Version})
	MATCH (p:Project {uuid: $projectUUID})
	MERGE (m)-[r:SCANNED]->(p)
	RETURN m, r, p;
	`

	parameters := map[string]interface{}{
		"Version":     version,
		"projectUUID": projectUUID,
	}

	_, err = session.Run(query, parameters)
	if err != nil {
		return fmt.Errorf("error creating SCANNED relationship: %s, %v", query, err)
	}

	logger.Debug("created SCANNED relationship", logger.LogFields{"version": version, "projectUUID": projectUUID})

	return nil
}

// LinkVolumeToServer creates a relationship between a volume and attached servers
func (r *Neo4jRepository) LinkVolumeToServer(volumeUUID string, serverID string) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	query := `
    MATCH (v:Volume {uuid: $volumeUUID}), (s:Server {id: $serverID})
    MERGE (v)-[:AttachedTo]->(s)
    `

	parameters := map[string]interface{}{
		"volumeUUID": volumeUUID,
		"serverID":   serverID,
	}

	_, err = session.Run(query, parameters)
	if err != nil {
		return fmt.Errorf("error creating AttachedTo relationship: %s, %v", query, err)
	}

	return nil
}

func (r *Neo4jRepository) LinkServerToProject(serverUUID string, projectID string) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	query := `
    MATCH (s:Server {uuid: $serverUUID}), (p:Project {id: $projectID})
    MERGE (s)-[:BelongsTo]->(p)
    `

	parameters := map[string]interface{}{
		"serverUUID": serverUUID,
		"projectID":  projectID,
	}

	_, err = session.Run(query, parameters)
	if err != nil {
		return fmt.Errorf("error creating BelongsTo relationship: %s, %v", query, err)
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
			MATCH (v:Volume {ID: $volumeID}), (s:Server {ID: $targetID})
			MERGE (v)-[:AttachedTo]->(s)
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
		logger.Error("Error creating Volume in Neo4j", err)
	} else {
		logger.Debug("Created volume in Neo4j", logger.LogFields{"volume_id": volume.ID})
	}
	return err
}
// CreateOrUpdateServer creates or updates a server node
func (r *Neo4jRepository) CreateOrUpdateServer(server dataparser.InfrastructureComponent) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	query := `
    MERGE (s:Server {ID: $id})
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
		"id":               server.ID,
		"name":             server.Name,
		"type":             server.Type,
		"availabilityZone": server.AvailabilityZone,
		"userID":           GetMetadataValue(server.Metadata, "UserID", ""),
		"hostID":           GetMetadataValue(server.Metadata, "HostID", ""),
		"tenantID":         GetMetadataValue(server.Metadata, "TenantID", ""),
		"created":          GetMetadataValue(server.Metadata, "Created", ""),
		"updated":          GetMetadataValue(server.Metadata, "Updated", ""),
		"volumesAttached":  GetMetadataValue(server.Metadata, "VolumesAttached", ""),
		"status":           GetMetadataValue(server.Metadata, "Status", ""),
	}

	_, err = session.Run(query, parameters)
	if err != nil {
		return fmt.Errorf("error creating Server in Neo4j: %v", err)
	} else { // if no err create relationship
		logger.Debug("Created server in Neo4j", logger.LogFields{"server_id": server.ID})

		// Handle relationships
		for _, rel := range server.Relationships {
			switch rel.Type {

			case "BelongsTo": // specific relationship type
				relationshipQuery := `
            MATCH (s:Server {ID: $serverID}), (p:Project {ID: $targetID})
            MERGE (s)-[:BelongsTo]->(p)
            `

				relationshipParameters := map[string]interface{}{
					"serverID": server.ID,
					"targetID": rel.Target,
				}

				_, err = session.Run(relationshipQuery, relationshipParameters)
				if err != nil {
					return err
				}

			case "AttachedTo":
				relationshipQuery := `
				MATCH (s:Server {ID: $serverID}), (v:Volume {ID: $volumeID})
				MERGE (s)-[:AttachedTo]->(v)
				`

				relationshipParameters := map[string]interface{}{
					"serverID": server.ID,
					"volumeID": rel.Target,
				}

				_, err = session.Run(relationshipQuery, relationshipParameters)
				if err != nil {
					return fmt.Errorf("error creating relationship between server and volume: %v", err)
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
            MATCH (p:Pod {ID: $podID}), (s:Server {Name: $serverName})
            MERGE (p)-[:RunsOn]->(s)
            `

				relationshipParameters := map[string]interface{}{
					"podID":      pod.ID,
					"serverName": rel.Target,
				}

				_, err = session.Run(relationshipQuery, relationshipParameters)
				if err != nil {
					return err
				}
			}
		}
		logger.Debug("Created pod in Neo4j", logger.LogFields{"pod_id": pod.ID})
	}
	return err
}

// FindInstanceByUUID finds a Instance by its uuid
func (r *Neo4jRepository) FindInstanceByUUID(ctx context.Context, uuid string) (*model.Instance, error) {
	query := `
		match (m:Instance) where m.uuid = $uuid return m.uuid, m.name, m.created, m.status
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
		logger.Error("Cannot find Instance by uuid", logger.LogFields{"uuid": uuid}, err)
	}

	logger.Debug("CYPHER_QUERY", logger.LogFields{"query": query, "args": args})

	Instance := model.Instance{}

	for result.Next() {
		ParseCypherQueryResult(result.Record(), "m", &Instance)
	}

	return &Instance, err
}

func (r *Neo4jRepository) TestNeo4jConnection(ctx context.Context) (string, error) {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return "", err
	}
	defer session.Close()

	result, err := session.Run("RETURN 1", nil)
	if err != nil {
		// Handle error, query execution failed
		fmt.Println("Failed to execute query:", err)
		return "", err
	}

	if result.Next() {
		record := result.Record()
		value := record.GetByIndex(0)
		// Process the fetched value as needed
		fmt.Println("Fetched Value:", value)
	}
	return "", err

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

	logger.Debug("CYPHER_QUERY", logger.LogFields{"query": query, "args": args})

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
