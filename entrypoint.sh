#!/bin/sh

# Wait for Neo4j to be ready
while ! curl -s http://neo4j:7474 > /dev/null; do
    echo "Waiting for Neo4j..."
    sleep 1
done

echo "Neo4j started!"

# Start the main application
./server
