## Why

Object and bucket ACLs are a fundamental GCS API feature used by SDKs for fine-grained access control. The models exist but there are no HTTP handlers, backend methods, or routes. Adding them closes one of the biggest remaining gaps in API coverage.

## What Changes

- Add ACL methods to Backend interface (GetObjectACL, UpdateObjectACL, GetBucketACL, UpdateBucketACL)
- Implement in MemoryBackend
- Add ACL handlers and routes
- Integration tests

## Impact

- `internal/backend/backend.go`: 4 new interface methods
- `internal/backend/memory.go`: ACL storage + methods
- `internal/model/`: ACL models (may already exist in bucket model)
- `internal/api/handler.go`: ACL handlers
- `internal/router/router.go`: ACL routes
