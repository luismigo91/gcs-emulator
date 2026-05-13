## 1. Router Refactor

- [x] 1.1 Add `RouterConfig` struct with named fields for all backends
- [x] 1.2 Implement `NewWithConfig(cfg RouterConfig) http.Handler`
- [x] 1.3 Rewrite `New()` as wrapper calling `NewWithConfig()`
- [x] 1.4 Update `pkg/emulator/emulator.go` to use `RouterConfig`
- [x] 1.5 Update `cmd/server/main.go` to use `RouterConfig`

## 2. Test Harness

- [x] 2.1 Create `TestHarness` struct with `*httptest.Server` and all backends
- [x] 2.2 Create `newEmulator(t, cfg RouterConfig) *TestHarness` factory
- [x] 2.3 Migrate existing test factories to use `RouterConfig` (backward compat)

## 3. Nivel 1: Table-Driven CRUD Tests

- [x] 3.1 Define `crudTestCase` struct: name, endpoint, createBody, entityName
- [x] 3.2 Write test table for Secret Manager
- [x] 3.3 Write test table for KMS
- [x] 3.4 Write test table for IAM Service Accounts
- [x] 3.5 Write test table for Cloud DNS
- [x] 3.6 Write test table for Cloud Billing
- [x] 3.7 Write test table for Service Directory
- [x] 3.8 Write test table for Artifact Registry
- [x] 3.9 Write test table for Cloud Build
- [x] 3.10 Write test table for CDN Backend Services
- [x] 3.11 Write test table for Cloud Scheduler

## 4. Nivel 2: Service-Specific Tests

- [x] 4.1 Cloud Tasks: test task dispatch, pause/resume lifecycle
- [x] 4.2 BigQuery: test SQL SELECT with WHERE clause and LIMIT
- [x] 4.3 Cloud Logging: test write entries + list back
- [x] 4.4 Cloud Monitoring: test time series ingest + query
- [x] 4.5 Cloud Trace: test span ingestion + trace query
- [x] 4.6 Error Reporting: test report error + list errors
- [x] 4.7 Pub/Sub DLQ: test dead letter routing after max attempts
- [x] 4.8 Pub/Sub Filter: test message filtering by attributes

## 5. Nivel 3: Infrastructure Smoke Tests

- [x] 5.1 Test health check reports all 19 services
- [x] 5.2 Test dashboard returns HTML 200 with service names
- [x] 5.3 Test /__/services lists all services as JSON
- [x] 5.4 Test /metrics returns Prometheus format
- [x] 5.5 Test /oauth2/v4/token returns valid token response

## 6. Existing Test Migration

- [x] 6.1 Update `newTestServer` to use `RouterConfig`
- [x] 6.2 Update `newPubSubTestServer` to use `RouterConfig`
- [x] 6.3 Update `newCloudTasksTestServer` to use `RouterConfig`
- [x] 6.4 Update `newSecretManagerTestServer` to use `RouterConfig`
- [x] 6.5 Update persistence test helpers to use `RouterConfig`

## 7. Validation

- [x] 7.1 Run `go test -race -coverprofile=coverage.out ./...`
- [x] 7.2 Verify coverage exceeds 80% for api packages
- [x] 7.3 Verify all existing tests still pass
- [x] 7.4 Run `go test -count=3 ./...` to verify test stability
