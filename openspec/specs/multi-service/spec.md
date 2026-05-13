# multi-service

## Requirements

### Requirement: Single Binary Architecture
The emulator SHALL serve all configured services from a single HTTP server on one port.

#### Scenario: Multiple services on port 9090
- **WHEN** emulator starts
- **THEN** all services (GCS, Pub/Sub, Secret Manager, etc.) are accessible on port 9090

### Requirement: Path-Based Routing
The emulator SHALL route requests by URL prefix to the appropriate service handler.

#### Scenario: GCS and Pub/Sub paths coexist
- **WHEN** GET /storage/v1/b and GET /v1/projects/{p}/topics
- **THEN** each request reaches the correct handler without interference

### Requirement: Health Check multi-service
The emulator SHALL report all running services in the health check response.

#### Scenario: Health reports all services
- **WHEN** GET /-/health
- **THEN** response includes status for all active services

### Requirement: Services Endpoint
The emulator SHALL expose an endpoint listing all active services.

#### Scenario: List services
- **WHEN** GET /__/services
- **THEN** JSON array of service names is returned

### Requirement: Shared Middleware
All services SHALL share project isolation, CORS, and auth passthrough middleware.

#### Scenario: Project header works across services
- **WHEN** X-Goog-User-Project header is sent with any service request
- **THEN** the request is scoped to that project

### Requirement: Per-Service Storage Configuration
The emulator SHALL support per-service storage mode and path overrides.

#### Scenario: Pub/Sub uses WAL while GCS uses memory
- **WHEN** GCP_EMULATOR_PUBSUB_MODE=wal is set
- **THEN** Pub/Sub uses WAL mode independently of GCS storage mode

### Requirement: Prometheus Metrics
The emulator SHALL expose Prometheus-compatible metrics for all services at /metrics.

#### Scenario: Metrics endpoint
- **WHEN** GET /metrics
- **THEN** Prometheus text format with request counts and durations is returned
