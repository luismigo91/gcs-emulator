## ADDED Requirements

### Requirement: Docker image distribution
The emulator SHALL be distributable as a Docker image that runs the emulator server on container start.

#### Scenario: Run from Docker image
- **WHEN** `docker run gcp-emulator/gcp-emulator:latest` is executed
- **THEN** the emulator starts and listens on port 9090 inside the container

#### Scenario: Port mapping
- **WHEN** the container is run with `-p 9090:9090`
- **THEN** the emulator is accessible at `http://localhost:9090` on the host machine

### Requirement: Configurable port
The emulator SHALL allow configuring the listening port via environment variable.

#### Scenario: Custom port
- **WHEN** `GCP_EMULATOR_PORT=8080` is set as an environment variable
- **THEN** the emulator listens on port 8080

#### Scenario: Default port
- **WHEN** `GCP_EMULATOR_PORT` is not set
- **THEN** the emulator listens on port 9090

### Requirement: Volume mount for data persistence
The emulator SHALL support mounting a host directory for data persistence when using persistent, hybrid, or WAL storage modes.

#### Scenario: Mount data directory
- **WHEN** the container is run with `-v $(pwd)/data:/app/data` and `GCP_EMULATOR_STORAGE_MODE=persistent`
- **THEN** data persists across container restarts in the host directory

### Requirement: Health check endpoint
The emulator SHALL expose a health check endpoint for Docker health checks and orchestration tools.

#### Scenario: Health check returns healthy
- **WHEN** a GET request is sent to `/-/health`
- **THEN** the emulator returns HTTP 200 with a JSON body `{"status": "healthy", "services": {"gcs": "available"}}`

#### Scenario: Docker HEALTHCHECK
- **WHEN** the Docker image includes a `HEALTHCHECK` instruction
- **THEN** Docker reports the container as healthy when the emulator is running

### Requirement: Docker Compose configuration
The emulator SHALL provide a ready-to-use `docker-compose.yml` example in the repository.

#### Scenario: Docker Compose quick start
- **WHEN** a user runs `docker compose up` with the provided compose file
- **THEN** the emulator starts with default configuration and port 9090 exposed

#### Scenario: Docker Compose with persistence
- **WHEN** the compose file includes a volume mount and `GCP_EMULATOR_STORAGE_MODE=persistent`
- **THEN** data persists across `docker compose down` and `docker compose up` cycles

### Requirement: Environment variable configuration
The emulator SHALL support all configuration via environment variables with `GCP_EMULATOR_` prefix.

#### Scenario: All settings via env vars
- **WHEN** `GCP_EMULATOR_PORT`, `GCP_EMULATOR_DEFAULT_PROJECT`, `GCP_EMULATOR_STORAGE_MODE`, and `GCP_EMULATOR_STORAGE_PATH` are set
- **THEN** the emulator uses all configured values without requiring CLI flags or config files

### Requirement: Graceful shutdown
The emulator SHALL handle SIGTERM and SIGINT signals for graceful shutdown, completing in-flight requests and flushing data.

#### Scenario: SIGTERM handling
- **WHEN** the emulator process receives SIGTERM
- **THEN** it stops accepting new requests, completes in-flight requests, flushes data (if applicable), and exits with code 0

#### Scenario: Docker stop
- **WHEN** `docker stop` is executed on a running emulator container
- **THEN** the emulator shuts down gracefully within the Docker stop timeout (default 10s)
