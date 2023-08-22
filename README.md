# Graph-builder service (Transparency Monitoring Service (TMS)) 
This repository holds the component for constructing the graph out of the different data sources. Planned architecture: 
[Excali Board](https://excalidraw.com/#json=nTY2HnHaiaMcYJOYK8beS,YWVmtXo6pRIJhX07fY_aPA)

## First Steps:
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

 ### Project Structure
```sh
.
│
├── cmd # TODO move server.go here, contains executables required to start app
├── deployment # deployment scripts for local or prod env
├── gqlgen.yml # GqlGen configuration
├── graph
│   │   
│   ├── generated # GqlGen generated files
│   ├── model # defines User and Todo model structs
│   └── resolver.go # implements the graphql queries/mutations
│    
├── schema.graphqls # graphql schema definition
│
├── internal # main components wiring applciation componemnts and
│   ├── dataparser   #  transforms provider plugin specfifc data into a generic model
│   ├── orchestrator #  TODO might move out of internal/ orchestrates registering plugins, data parsing and transfroming
│   ├── plugin       #  provides functionality to register, inialize provider specific instances and fetch Data
│   ├── repository   #  adpater for database connection
│   └── services     #  service layer, exposes busniss logic to the application define operations executed from             │                       external interfaces such as graphQL endpoints
│ 
├── pkg
│   └── logger #  
│      ├── logger.go # logger interface
│      └── globals.go # ...
└── server.go # runnable server
```

## Generate code gqlgen 
```sh
# Update generated Code 
go generate ./...
```

## Import data
```sh
# Import dummy data into database
docker-compose exec neo4j /bin/bash -c 'cat /var/lib/neo4j/import/resource.cyhpher | cypher-shell -u neo4j -p testingshit'
```

## Run service locally
```sh
# start docker container 
cd deployments
docker compose up 
go run ./server.go 
# Access GraphQL playground
http://192.168.2.139:8080/playground
# Access Neo4j Database UI
http://localhost:7474/browser/
```

## Visualizing code
Requirement install go-callvis and Graphviz (https://www.graphviz.org/download/)

```sh
go install github.com/ofabry/go-callvis@latest
brew install graphviz 
# To generate graph run:
 go-callvis github.com/regulatory-transparency-monitor/graph-builder
```