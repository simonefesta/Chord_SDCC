#!/bin/bash

x=${1:-0}  # If no argument is provided, set x to 0 by default

# Read the value of 'm' from config.json
m=$(cat config.json | jq -r '.bits')

# Calculate the port based on '8005 + m + x'
node_number=$((m + x +1 ))

port=$((8005 + node_number))

# Build the Docker image
docker build -t "chord_sdcc_node${node_number}" -f DockerFiles/node/Dockerfile .

# Run the Docker container with the calculated port
docker run -p $port:8005 --network=chord_sdcc_my_network --name="chord_sdcc_node${node_number}" "chord_sdcc_node${node_number}"
