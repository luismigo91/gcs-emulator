## Why

14 of 19 services have 0% test coverage. The existing test infrastructure is solid (httptest.Server + doRequest pattern) but creating test servers requires passing 7+ positional params to `router.New`, making each new test factory brittle and inconsistent. The service handlers follow identical CRUD patterns, making them ideal for table-driven testing.

## What Changes

### Router ergonomics
- Add `RouterConfig` struct to replace the 7-param `router.New()` signature
- Backward-compatible: existing `New()` delegates to `NewWithConfig(RouterConfig{...})`
- Each test factory only sets the backends it needs

### Table-driven CRUD tests (Nivel 1)
- Single test file that covers 10 CRUD-heavy services
- Each service entry specifies: endpoint prefix, create body, create path, entity name for get/delete
- Covers: Secret Manager, KMS, IAM, DNS, Billing, Service Directory, Artifact Registry, Cloud Build, CDN, Scheduler

### Service-specific tests (Nivel 2)
- Cloud Tasks: task dispatch, pause/resume lifecycle
- BigQuery: SQL SELECT with WHERE, LIMIT, column projection
- Logging/Monitoring/Trace/Error Reporting: write + query roundtrip
- Pub/Sub: published messages reach subscribers, DLQ routing, message filtering

### Infrastructure smoke tests (Nivel 3)
- Health check reports all services
- Dashboard returns HTML 200
- Services endpoint lists all services
- Metrics endpoint exposes Prometheus format
- Auth token endpoint returns valid response
- Persistence roundtrip for each storage backend

## Capabilities

### Modified Capabilities
- `multi-service`: Router uses `RouterConfig` struct instead of positional params

## Impact

- `internal/router/router.go`: Add `RouterConfig`, `NewWithConfig()`, keep `New()` as wrapper
- `internal/api/*_test.go`: New test files for each untested service
- Existing test factories: Updated to use `RouterConfig`
