# Transparency Monitoring Service (TMS)
This repository holds the component for constructing the graph out of the different data sources. Planned architecture: 
[Excali Board](https://excalidraw.com/#json=nTY2HnHaiaMcYJOYK8beS,YWVmtXo6pRIJhX07fY_aPA)

## Local Deployment: 
```sh
# start docker container 
cd deployments
docker compose up 
# Access GraphQL API playground
http://0.0.0.0:8080/playground
# Access Neo4j Database UI
http://localhost:7474/browser/
# Access Transparency Dashboard
http://localhost:5005/ 
```
```sh
# To rebuild image:
docker build -t transparency-monitor -f deployments/Dockerfile .
```

### Development mode
```sh
# A running neo4j container is necessary 
cd cmd 
go run ./server.go 
```

#
## Import data
```sh
# Import dummy data into database
docker-compose exec neo4j /bin/bash -c 'cat /var/lib/neo4j/import/resource.cyhpher | cypher-shell -u neo4j -p psw'
```


 ### Project Structure
```sh
.
│
├── cmd/            # Contains executables required to start app
│   └── server.go   # runnable server
├── deployment      # Deployment scripts 
│   ├── terrarform/ # Deployment script for OpenStack
│   ├── docker-compose.yml  # Docker compose
│   ├── Dockerfile  # Building monitoring serivce image
│   └── deploy-script.sh    # Prepare VM
├── gqlgen.yml      # GqlGen configuration
├── graph/  # GraphQL API
│   ├── generated   # GqlGen generated files
│   ├── model       # Defines User and  model structs
│   └── resolver.go # Implements the graphql queries/mutations
│    
├── schema.graphqls # Graphql schema definition
│
├── internal/
│   └── app.go      # Glues together the application logic (internal graphQL server, neo4j database connection, monitoring workflow)           
│   ├── manager/     # Manages monitoring workflow
│   │   ├── manager.go          # Orchestrates and schedules workflow phases
│   │   └── scheduler.go        # Scheduler job
│   ├── repository/  # Neo4j DB repository
│   │   ├── interface.go        # Repository definitions
│   │   ├── neo4j.go            # Cypher functions to create/update nodes 
│   │   ├── relationships.go    # Cypher functions to create rel
│   │   └── utils.go            # Helper functions
│   ├── service/     # Service layer exposes bussiness logic  
│   │   └── service.go      
│   └── versioning  # Handles graph versioning
│       └── versioning.go   
│                           
├── pkg/
│   ├── dataparser/  # Transfromation phase
│   │    ├── genericModel.go            # Model of the gernic data types 
│   │    ├── transformer.go             # Transformer interface
│   │    ├── kubernetes_transformer.go  # Custom data mapper, applies kubernetes domain knowledge
│   │    └── openstack_transformer.go   # Custom data mapper, applies openstack domain knowledge
│   ├── logger/      # Service Logger
│   │    └── logger.go                  # Logger interface
│   │    └── globals.go 
│   └── plugin/      # Data Collection Phase
│        ├── pluginManager.go           # Plugin interface
│        ├── registry.go                # Register active plugins 
│        └── scanner.go                 # Fetches data from data sources

```




## Visualizing code
Requirement install go-callvis and Graphviz (https://www.graphviz.org/download/)

```sh
go install github.com/ofabry/go-callvis@latest
brew install graphviz 
# To generate graph run:
 go-callvis github.com/regulatory-transparency-monitor/graph-builder
```
## Documentation
```sh
# Create a directory for project and initialize go modules file:
go mod init github.com/regulatory-transparency-monitor/graph-builder
```

```sh
# Use ‍‍gqlgen init command to setup a gqlgen project: 
go run github.com/99designs/gqlgen init
```

```sh
# Defining schema in **file: graph/schema.graphqls** and run:
go run github.com/99designs/gqlgen generate
```
### Generate code gqlgen for GraphQL API 
```sh
# Update generated functions  
go generate ./...
```