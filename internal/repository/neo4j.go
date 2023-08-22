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

// NewNeo4jConnection creates a new neo4j connection
// returns a neo4j.Driver object and an error
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

// Neo4jRepository is a Neo4j DB repository
type Neo4jRepository struct {
	Connection neo4j.Driver
}

// CreateOrUpdateProject creates or updates a  project node
func (r *Neo4jRepository) CreateOrUpdateProject(project dataparser.InfrastructureComponent) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	query := `
    MERGE (p:Project {ID: $id})
    ON CREATE SET p.Name = $name, p.Type = $type, p.AvailabilityZone = $availabilityZone, p.Enabled = $enabled, p.Description = $description
    ON MATCH SET p.Name = $name, p.Type = $type, p.AvailabilityZone = $availabilityZone, p.Enabled = $enabled, p.Description = $description
    `

	parameters := map[string]interface{}{
		"id":               project.ID,
		"name":             project.Name,
		"type":             project.Type,
		"availabilityZone": project.AvailabilityZone,
		"description":      GetMetadataValue(project.Metadata, "Description", ""),
		"enabled":          GetMetadataValue(project.Metadata, "Enabled", false),
	}

	_, err = session.Run(query, parameters)
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
		logger.Error("Error creating server in Neo4j", err)
	} else { // if no err create relationship
		logger.Debug("Created server in Neo4j", logger.LogFields{"server_id": server.ID})

		// Handle relationships
		for _, rel := range server.Relationships {
			if rel.Type == "BelongsTo" { // specific relationship type
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
			}

		}

	}
	return err
}

// CreateOrUpdateVolume creates or updates a Volume Node
func (r *Neo4jRepository) CreateOrUpdateVolume(project dataparser.InfrastructureComponent) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	query := `
    MERGE (s:Server {ID: $id})
    ON CREATE SET p.Name = $name, p.Type = $type, p.AvailabilityZone = $availabilityZone
    ON MATCH SET p.Name = $name, p.Type = $type, p.AvailabilityZone = $availabilityZone
    `

	parameters := map[string]interface{}{
		"id":               project.ID,
		"name":             project.Name,
		"type":             project.Type,
		"availabilityZone": project.AvailabilityZone,
		"metadata":         project.Metadata,
	}

	_, err = session.Run(query, parameters)
	return err
}

// CreateOrUpdatePod creates or updates a Kubernetes Pod
func (r *Neo4jRepository) CreateOrUpdatePod(project dataparser.InfrastructureComponent) error {
	session, err := r.Connection.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer session.Close()

	query := `
    MERGE (s:Server {ID: $id})
    ON CREATE SET p.Name = $name, p.Type = $type, p.AvailabilityZone = $availabilityZone
    ON MATCH SET p.Name = $name, p.Type = $type, p.AvailabilityZone = $availabilityZone
    `

	parameters := map[string]interface{}{
		"id":               project.ID,
		"name":             project.Name,
		"type":             project.Type,
		"availabilityZone": project.AvailabilityZone,
		"metadata":         project.Metadata,
	}

	_, err = session.Run(query, parameters)
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
