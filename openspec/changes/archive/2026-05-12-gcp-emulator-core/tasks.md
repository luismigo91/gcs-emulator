## 1. Project Setup

- [x] 1.1 Initialize Go module with `go mod init`
- [x] 1.2 Create directory structure: `cmd/server/`, `internal/api/`, `internal/backend/`, `internal/model/`, `internal/router/`, `internal/auth/`, `pkg/emulator/`, `tests/`
- [x] 1.3 Add dependencies: `google.golang.org/api` for GCS types, `github.com/google/uuid` for object IDs
- [x] 1.4 Create `Makefile` with targets: `build`, `test`, `run`, `docker-build`, `lint`
- [x] 1.5 Create initial `Dockerfile` (multi-stage: Go build → scratch/alpine)
- [x] 1.6 Create `docker-compose.yml` with default configuration
- [x] 1.7 Set up `.gitignore` for Go projects

## 2. Domain Models

- [x] 2.1 Define `Bucket` struct with all GCS bucket fields (name, projectNumber, metageneration, versioning, lifecycle, iamConfiguration, etc.)
- [x] 2.2 Define `Object` struct with all GCS object fields (name, bucket, generation, metageneration, contentType, size, md5Hash, crc32c, metadata, etc.)
- [x] 2.3 Define `ObjectAttrs` and `BucketAttrs` for create/update operations
- [x] 2.4 Define `Project` model for project-scoped resource grouping
- [x] 2.5 Define GCP error response types matching the JSON error format

## 3. Storage Backend Interface

- [x] 3.1 Define `Backend` interface in `internal/backend/backend.go` with all CRUD methods
- [x] 3.2 Define error types: `ErrBucketNotFound`, `ErrObjectNotFound`, `ErrBucketNotEmpty`, `ErrGenerationNotFound`
- [x] 3.3 Implement `MemoryBackend` with `map[string]*Bucket` and `map[string]*Object` + `sync.RWMutex`
- [x] 3.4 Implement `PersistentBackend` extending MemoryBackend with JSON save/load on shutdown/startup
- [x] 3.5 Implement `HybridBackend` with async flush goroutine (5s interval)
- [x] 3.6 Implement `WALBackend` with append-only JSON lines log and replay on startup
- [x] 3.7 Implement WAL compaction (snapshot + truncate at 10MB or on shutdown)
- [x] 3.8 Implement backend factory function that selects backend based on config

## 4. Configuration

- [x] 4.1 Define `Config` struct with all configuration fields (Port, DefaultProject, StorageMode, StoragePath)
- [x] 4.2 Implement environment variable loading with `GCP_EMULATOR_` prefix and defaults
- [x] 4.3 Implement per-service storage mode override parsing

## 5. HTTP Router

- [x] 5.1 Create router setup in `internal/router/router.go` using `http.ServeMux`
- [x] 5.2 Register JSON API routes: `/storage/v1/b`, `/storage/v1/b/{bucket}`, `/storage/v1/b/{bucket}/o`, `/storage/v1/b/{bucket}/o/{object}`
- [x] 5.3 Register upload routes: `/upload/storage/v1/b/{bucket}/o`, `/resumable/upload/storage/v1/b/{bucket}/o`
- [x] 5.4 Register XML API routes: `/{bucket}`, `/{bucket}/{object}`
- [x] 5.5 Register health endpoint: `/-/health`
- [x] 5.6 Implement request context middleware for project ID resolution
- [x] 5.7 Implement CORS headers for browser-based SDK usage

## 6. GCS API Handlers — Buckets

- [x] 6.1 Implement `POST /storage/v1/b` — Create bucket
- [x] 6.2 Implement `GET /storage/v1/b` — List buckets (with project filter)
- [x] 6.3 Implement `GET /storage/v1/b/{bucket}` — Get bucket metadata
- [x] 6.4 Implement `DELETE /storage/v1/b/{bucket}` — Delete bucket (empty only)
- [x] 6.5 Implement `PATCH /storage/v1/b/{bucket}` — Update bucket (versioning, lifecycle, etc.)

## 7. GCS API Handlers — Objects

- [x] 7.1 Implement `POST /storage/v1/b/{bucket}/o` — Create object (multipart upload)
- [x] 7.2 Implement `GET /storage/v1/b/{bucket}/o` — List objects (with prefix, delimiter, maxResults)
- [x] 7.3 Implement `GET /storage/v1/b/{bucket}/o/{object}` — Get object metadata
- [x] 7.4 Implement `GET /storage/v1/b/{bucket}/o/{object}?alt=media` — Download object content
- [x] 7.5 Implement `DELETE /storage/v1/b/{bucket}/o/{object}` — Delete object
- [x] 7.6 Implement `PATCH /storage/v1/b/{bucket}/o/{object}` — Update object metadata
- [x] 7.7 Implement `POST /storage/v1/b/{bucket}/o/{object}/compose` — Compose objects
- [x] 7.8 Implement `POST /storage/v1/b/{bucket}/o/{sourceObject}/copyTo/b/{destBucket}/o/{destObject}` — Copy object
- [x] 7.9 Implement `POST /storage/v1/b/{bucket}/o/{sourceObject}/rewriteTo/b/{destBucket}/o/{destObject}` — Rewrite object
- [x] 7.10 Implement Range header support for partial downloads (HTTP 206)
- [x] 7.11 Implement generation parameter support for versioned objects

## 8. Resumable Upload

- [x] 8.1 Implement resumable upload session creation (POST with `X-Goog-Upload-Command: start`)
- [x] 8.2 Implement chunk upload handler (PUT with `Content-Range` header)
- [x] 8.3 Implement upload finalization (`X-Goog-Upload-Command: upload, finalize`)
- [x] 8.4 Implement upload status query (`X-Goog-Upload-Command: query`)
- [x] 8.5 Implement session cleanup on completion or expiration

## 9. XML API Handlers

- [x] 9.1 Implement `PUT /{bucket}/{object}` — Upload object (simple)
- [x] 9.2 Implement `GET /{bucket}/{object}` — Download object
- [x] 9.3 Implement `DELETE /{bucket}/{object}` — Delete object
- [x] 9.4 Implement `GET /{bucket}` — List objects (XML format)
- [x] 9.5 Implement `PUT /{bucket}` — Create bucket

## 10. Bucket Lifecycle Rules

- [x] 10.1 Implement lifecycle rule storage on bucket configuration
- [x] 10.2 Implement lifecycle rule evaluation engine (age, createdBefore, numNewerVersions conditions)
- [x] 10.3 Implement lifecycle actions (Delete, SetStorageClass)
- [x] 10.4 Integrate lifecycle evaluation into object listing and retrieval

## 11. Bucket IAM Policies

- [x] 11.1 Implement `GET /storage/v1/b/{bucket}/iam` — Get IAM policy
- [x] 11.2 Implement `PUT /storage/v1/b/{bucket}/iam` — Set IAM policy
- [x] 11.3 Implement `POST /storage/v1/b/{bucket}/iam/testPermissions` — Test IAM permissions (stub)

## 12. Bucket Notifications

- [x] 12.1 Implement `POST /storage/v1/b/{bucket}/notificationConfigs` — Create notification
- [x] 12.2 Implement `GET /storage/v1/b/{bucket}/notificationConfigs` — List notifications
- [x] 12.3 Implement `GET /storage/v1/b/{bucket}/notificationConfigs/{id}` — Get notification
- [x] 12.4 Implement `DELETE /storage/v1/b/{bucket}/notificationConfigs/{id}` — Delete notification

## 13. Auth Emulation

- [x] 13.1 Implement OAuth2 token endpoint returning mock tokens
- [x] 13.2 Implement auth middleware that accepts any credentials (pass-through)
- [x] 13.3 Implement signed URL passthrough (ignore signature validation)
- [x] 13.4 Ensure `STORAGE_EMULATOR_HOST` and `option.WithoutAuthentication()` work with Go SDK

## 14. Error Handling

- [x] 14.1 Implement GCP JSON error response formatter
- [x] 14.2 Map Go errors to GCP error codes (400, 403, 404, 409, 412, 416, 500)
- [x] 14.3 Implement error response for all handler endpoints
- [x] 14.4 Implement XML API error responses

## 15. Public Go Library API

- [x] 15.1 Create `pkg/emulator/emulator.go` with `Server` struct
- [x] 15.2 Implement `NewServer(config Config) (*Server, error)` constructor
- [x] 15.3 Implement `Server.Start()` and `Server.Stop()` methods
- [x] 15.4 Implement `Server.URL()` for test integration
- [x] 15.5 Add Go library usage examples in `pkg/emulator/example_test.go`

## 16. CLI Entry Point

- [x] 16.1 Create `cmd/server/main.go` with CLI entry point
- [x] 16.2 Implement graceful shutdown handling (SIGTERM, SIGINT)
- [x] 16.3 Implement startup logging with configuration summary
- [x] 16.4 Build binary with `go build`

## 17. Docker Configuration

- [x] 17.1 Finalize multi-stage `Dockerfile` (Go build → alpine, ~15-20MB image)
- [x] 17.2 Add `HEALTHCHECK` instruction using `/-/health` endpoint
- [x] 17.3 Create production-ready `docker-compose.yml` with volume mount example
- [x] 17.4 Create `.dockerignore` file
- [x] 17.5 Test container startup, port mapping, and volume persistence

## 18. Testing

- [x] 18.1 Write unit tests for all storage backend modes
- [x] 18.2 Write unit tests for domain models and serialization
- [ ] 18.3 Write integration tests for all bucket API endpoints
- [ ] 18.4 Write integration tests for all object API endpoints
- [ ] 18.5 Write integration tests for resumable uploads
- [ ] 18.6 Write integration tests for XML API endpoints
- [ ] 18.7 Write integration tests for versioning and lifecycle
- [ ] 18.8 Write integration tests for IAM and notifications
- [ ] 18.9 Write concurrent access tests for race conditions
- [ ] 19.0 Write graceful shutdown tests

## 19. Compatibility Tests

- [ ] 19.1 Set up Go SDK compatibility tests using `cloud.google.com/go/storage`
- [ ] 19.2 Set up Python SDK compatibility tests using `google-cloud-storage`
- [ ] 19.3 Set up Node.js SDK compatibility tests using `@google-cloud/storage`
- [ ] 19.4 Run compatibility test suite and document passing/failing endpoints

## 20. Documentation

- [x] 20.1 Update `README.md` with accurate project scope, features, and quick start
- [x] 20.2 Add SDK usage examples (Go, Python, Node.js, Java)
- [x] 20.3 Add configuration reference documentation
- [x] 20.4 Add contributing guidelines
- [ ] 20.5 Add architecture overview diagram
