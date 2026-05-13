## MODIFIED Requirements

### Requirement: Health check endpoint
The emulator SHALL expose a health check endpoint for Docker health checks and orchestration tools.

#### Scenario: Health check returns healthy
- **WHEN** a GET request is sent to `/-/health`
- **THEN** the emulator returns HTTP 200 with a JSON body `{"status":"healthy","services":{"gcs":"available"}}`

#### Scenario: Docker HEALTHCHECK
- **WHEN** the Docker image includes a `HEALTHCHECK` instruction
- **THEN** Docker reports the container as healthy when the emulator is running
