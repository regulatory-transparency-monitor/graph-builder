# Graph-builder service (Transparency Monitoring Service (TMS)) 
This repository holds the component for constructing the graph out of the different data sources. 


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
├── cmd # contains executables required to start app
├── deployment # deployment scripts for local or prod env
├── gqlgen.yml # GqlGen configuration
├── graph
│   ├── dataloader # TODO implements the user dataloader 
│   ├── generated # GqlGen generated files
│   ├── model # defines User and Todo model structs
│   └── resolver.go # implements the graphql queries/mutations
│    
├── schema.graphqls # graphql schema definition
│
├── internal
│   ├── db # TODO 
│   └── services # TODO
├── pkg
│   └── logger # TODO 
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


## Visualizing code
Requirement install go-callvis and Graphviz (https://www.graphviz.org/download/)

```sh
go install github.com/ofabry/go-callvis@latest
brew install graphviz 
# To generate graph run:
 go-callvis github.com/regulatory-transparency-monitor/graph-builder
```