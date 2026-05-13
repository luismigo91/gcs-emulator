## MODIFIED Requirements

### Requirement: Docker image distribution
The emulator SHALL be distributable as a Docker image that runs both GCS and Pub/Sub services on container start.

#### Scenario: Run from Docker image
- **WHEN** `docker run gcs-emulator/gcs-emulator:latest` is executed
- **THEN** the emulator starts and listens on port 9090 serving both GCS and Pub/Sub APIs

### Requirement: Health check endpoint
The emulator SHALL expose a health check endpoint reporting status of all running services.

#### Scenario: Health check returns both services
- **WHEN** a GET request is sent to `/-/health`
- **THEN** the emulator returns HTTP 200 with JSON body `{"status":"healthy","services":{"gcs":"available","pubsub":"available"}}`
