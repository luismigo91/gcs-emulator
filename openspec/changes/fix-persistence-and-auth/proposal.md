## Why

The persistent, hybrid, and WAL storage backends lose object binary content, IAM policies, and notification configurations on restart — making non-memory modes unusable for any real workflow. Meanwhile, the `internal/auth/` directory is completely empty despite the auth-emulation spec defining an OAuth2 token endpoint required by several SDKs for their credential flow.

## What Changes

### Data Persistence
- **Persist object binary content** in Persistent, Hybrid, and WAL backends (currently only metadata is saved; content resets to zeroed buffers)
- **Persist IAM policies** across restarts in disk-backed modes
- **Persist notification configurations** across restarts in disk-backed modes

### Auth Emulation
- **Implement `internal/auth/` package** with an OAuth2 token endpoint (`/oauth2/v4/token`) that returns valid-looking tokens
- **Register auth routes** in the router

## Capabilities

### New Capabilities
- `auth-token-endpoint`: OAuth2 token endpoint that returns mock tokens for SDK credential flow compatibility

### Modified Capabilities
- `storage-backend`: Extend persistence to include binary content, IAM policies, and notification configurations (previously lost on restart)

## Impact

- `internal/auth/`: New package implementing OAuth2 token endpoint
- `internal/backend/persistent.go`: Save/load binary content and IAM/notification data
- `internal/backend/hybrid.go`: Same persistence fixes
- `internal/backend/wal.go`: WAL replay/compaction to include IAM and notification operations
- `internal/router/router.go`: Register `/oauth2/v4/token` route
- `internal/model/`: No changes needed (models already exist)
