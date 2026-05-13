## ADDED Requirements

### Requirement: Multi-service HTTP routing
The emulator SHALL serve both GCS and Pub/Sub APIs from a single HTTP server, routing requests by URL path prefix.

#### Scenario: GCS routes still work
- **WHEN** a request to `/storage/v1/b` is sent
- **THEN** the GCS bucket handler processes it

#### Scenario: Pub/Sub routes work
- **WHEN** a request to `/v1/projects/test-project/topics/my-topic` is sent
- **THEN** the Pub/Sub handler processes it

#### Scenario: Health check reports both services
- **WHEN** `GET /-/health` is called
- **THEN** the response includes `{"services":{"gcs":"available","pubsub":"available"}}`

### Requirement: Service-level config overrides
The emulator SHALL allow per-service storage mode and path overrides via environment variables.

#### Scenario: Pub/Sub uses WAL while GCS uses memory
- **WHEN** `GCP_EMULATOR_SERVICE_PUBSUB=mode:wal,path:/data/pubsub` is set
- **THEN** Pub/Sub uses WAL mode at `/data/pubsub` while GCS uses the global default

### Requirement: Shared middleware
All emulator services SHALL share project isolation middleware, CORS headers, and auth passthrough.

#### Scenario: Project header works for Pub/Sub
- **WHEN** `X-Goog-User-Project: my-project` header is sent with a Pub/Sub request
- **THEN** the request is scoped to `my-project`
