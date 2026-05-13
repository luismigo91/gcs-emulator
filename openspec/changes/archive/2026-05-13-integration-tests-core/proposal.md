## Why

The GCS emulator has core API implementation and SDK compatibility features, but lacks integration tests that validate the full HTTP API surface against real SDK behavior. Without these tests, we cannot guarantee that Go, Python, and Node.js SDKs will work correctly. Additionally, three critical features are missing that prevent basic SDK usage: HEAD requests (used by SDKs for existence checks), data preloading (essential for test fixtures), and concurrent access safety.

## What Changes

- **Integration Test Suite**: HTTP-level tests for all bucket and object API endpoints, validating request/response shapes match real GCS
- **SDK Compatibility Tests**: Test scripts using `cloud.google.com/go/storage`, `google-cloud-storage` (Python), and `@google-cloud/storage` (Node.js) against the running emulator
- **HEAD Request Support**: `HEAD /storage/v1/b/{bucket}` and `HEAD /storage/v1/b/{bucket}/o/{object}` for existence checks
- **Data Preloading**: Seed directory support (`/data`) to preload buckets and objects on startup, matching fake-gcs-server behavior
- **Concurrent Access Tests**: Race condition detection with `go test -race`
- **Graceful Shutdown Tests**: Verify backend persistence on SIGTERM/SIGINT

## Capabilities

### Modified Capabilities

- `gcs-api`: Extended to include HEAD request support for bucket and object existence checks
- `docker-distribution`: Extended to include data preloading via volume mount at `/data`

## Impact

- `internal/api`: Add HEAD request handlers
- `internal/backend`: Add seed/preload functionality
- `cmd/server`: Add data preloading on startup
- `tests/`: New integration test files
- `pkg/emulator`: Add seed directory option to server constructor
