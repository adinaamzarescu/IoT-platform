#!/bin/bash

# Exit if any command fails
set -e

# Initialize Docker Swarm if it's not already active
if ! docker info 2>/dev/null | grep -q 'Swarm: active'; then
  echo "Initializing Docker Swarm..."
  docker swarm init
else
  echo "Docker Swarm is already active."
fi

# Deploy the stack
echo "Deploying stack..."
docker stack deploy -c stack.yml sprc

echo "Deployment complete. Services are starting up."
