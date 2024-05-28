# Exit if any command fails
$ErrorActionPreference = "Stop"

# Function to check if Docker Swarm is active
function Check-SwarmActive {
    try {
        $swarmStatus = docker info | Select-String "Swarm: active"
        return $swarmStatus -ne $null
    } catch {
        return $false
    }
}

# Initialize Docker Swarm if it's not already active
if (-not (Check-SwarmActive)) {
    Write-Host "Initializing Docker Swarm..."
    docker swarm init
} else {
    Write-Host "Docker Swarm is already active."
}

# Deploy the stack
Write-Host "Deploying stack..."
docker stack deploy -c stack.yaml sprc

Write-Host "Deployment complete. Services are starting up."
