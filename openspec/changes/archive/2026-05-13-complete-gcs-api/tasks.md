## 1. Lifecycle Rules

- [x] 1.1 Add `GetLifecycle`, `SetLifecycle`, `DeleteLifecycle` methods to Backend interface
- [x] 1.2 Implement lifecycle methods in MemoryBackend (delegate to bucket.Lifecycle field)
- [x] 1.3 Add lifecycle handler methods to `internal/api/handler.go`
- [x] 1.4 Register lifecycle routes: GET/PATCH/DELETE `/storage/v1/b/{bucket}/lifecycle`
- [x] 1.5 Add lifecycle integration tests

## 2. CORS Configuration

- [x] 2.1 Add `GetCORS`, `SetCORS` methods to Backend interface
- [x] 2.2 Implement CORS methods in MemoryBackend (delegate to bucket.CORS field)
- [x] 2.3 Add CORS handler methods to `internal/api/handler.go`
- [x] 2.4 Register CORS routes: GET/PATCH `/storage/v1/b/{bucket}/cors`
- [x] 2.5 Add CORS integration tests

## 3. XML API

- [x] 3.1 Add DELETE bucket case to XML API handler in `internal/api/handler.go`
- [x] 3.2 Add XML DELETE bucket integration test

## 4. Health Check

- [x] 4.1 Update HealthHandler response format to `{"status":"healthy","services":{"gcs":"available"}}`
- [x] 4.2 Update health check integration test

## 5. Validation

- [x] 5.1 Run `go test -race ./...` and confirm all passes
- [x] 5.2 Run `make lint`
