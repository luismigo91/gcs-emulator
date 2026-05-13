## Why

Several GCS API features are partially or not implemented: lifecycle rules (models exist, no API handlers), CORS configuration per bucket (models exist, no API), XML API DELETE bucket (missing), and the health check response format doesn't match the spec. These gaps prevent full SDK compatibility for bucket management operations.

## What Changes

### Lifecycle Rules
- Add HTTP handlers for bucket lifecycle CRUD (GET/PATCH/DELETE)
- Add lifecycle storage methods to the Backend interface and MemoryBackend
- Register lifecycle routes in the router

### CORS Configuration
- Add HTTP handlers for bucket CORS CRUD (GET/PATCH)
- Add CORS storage methods to the Backend interface and MemoryBackend
- Register CORS routes in the router

### XML API
- Add DELETE bucket endpoint to XML API handler (`DELETE /{bucket}`)

### Health Check
- Align response format with docker-distribution spec: `{"status":"healthy","services":{"gcs":"available"}}` instead of `{"status":"ok"}`

## Capabilities

### New Capabilities
- `lifecycle-rules`: Bucket lifecycle configuration management API (GET/PATCH/DELETE)
- `cors-configuration`: Bucket CORS rules management API (GET/PATCH)

### Modified Capabilities
- `gcs-api`: XML API now supports DELETE bucket; health check response format updated
- `docker-distribution`: Health check endpoint response format aligned with spec

## Impact

- `internal/api/handler.go`: New lifecycle and CORS handlers, XML DELETE bucket, updated health handler
- `internal/model/bucket.go`: No changes needed (models already exist)
- `internal/backend/backend.go`: Add lifecycle and CORS methods to Backend interface
- `internal/backend/memory.go`: Add lifecycle and CORS storage
- `internal/router/router.go`: Register lifecycle and CORS routes
