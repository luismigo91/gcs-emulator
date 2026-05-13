## Context

The GCS emulator currently supports core bucket/object operations, uploads, versioning, and multiple storage backends. However, several features expected by production GCS clients are missing or incomplete. This change systematically addresses all gaps identified during the integration testing phase.

## Goals / Non-Goals

**Goals:**
- Complete all remaining technical debt items
- Add IAM, notifications, signed URLs, lifecycle, and CORS support
- Verify Python and Node.js SDK compatibility
- Add performance benchmarks and metrics
- Optimize Docker image size

**Non-Goals:**
- gRPC API support
- Real Pub/Sub integration (notifications are stored, not delivered)
- Actual GCP authentication (still accept any auth)

## Decisions

### ETag: Use generation-based hash
**Decision**: Set ETag to base64-encoded generation number.
**Rationale**: Simple, deterministic, matches GCS behavior closely enough for SDK compatibility.

### IAM: Simplified policy model
**Decision**: Store IAM policies as JSON with basic role/member structure. No actual permission enforcement.
**Rationale**: SDKs expect the API to exist and return valid responses. Full enforcement is out of scope.

### Notifications: Storage-only
**Decision**: Store notification configurations but don't actually deliver messages.
**Rationale**: Emulator can't run Pub/Sub. SDKs only need to verify configuration CRUD.

### Signed URLs: Validation-only
**Decision**: Generate signed URLs that can be validated (signature check), but also accept any signed URL for emulator convenience.
**Rationale**: Real signature validation requires service account keys. For local dev, accepting any URL is more practical.

### Metrics: In-memory counters
**Decision**: Use sync/atomic counters for request metrics, expose via /metrics endpoint in Prometheus format.
**Rationale**: No external dependencies needed. Simple and effective for local development.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Large change scope | Group related changes, test incrementally |
| SDK compatibility tests may fail | Document known failures, fix critical ones |
| Docker optimization may break build | Test Docker build in CI |
