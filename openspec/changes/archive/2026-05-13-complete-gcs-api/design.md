## Context

Lifecycle rules and CORS configuration models already exist in `internal/model/bucket.go` (`Lifecycle`, `LifecycleRule`, `CORSRule`). The `BucketUpdateAttrs` struct includes both fields. The `UpdateBucket` method in MemoryBackend already handles lifecycle and CORS updates. What's missing is the HTTP API surface and backend interface methods dedicated to these sub-resources.

The GCS API supports:
- `GET /storage/v1/b/{bucket}/lifecycle` — get lifecycle configuration
- `PATCH /storage/v1/b/{bucket}/lifecycle` — set/update lifecycle configuration
- `DELETE /storage/v1/b/{bucket}/lifecycle` — remove lifecycle configuration
- `GET /storage/v1/b/{bucket}/cors` — get CORS configuration
- `PATCH /storage/v1/b/{bucket}/cors` — set/update CORS configuration

Currently, lifecycle and CORS are only modifiable via bucket-level `PATCH /storage/v1/b/{bucket}` with full `BucketUpdateAttrs`. The dedicated sub-resource endpoints aren't exposed.

The XML API supports `DELETE /{bucket}` but the handler only routes DELETE to object deletion — bucket deletion via XML is rejected as method not allowed.

## Goals / Non-Goals

**Goals:**
- Add dedicated lifecycle and CORS sub-resource endpoints matching GCS API
- Add XML API DELETE bucket support
- Align health check response with docker-distribution spec
- Follow existing handler patterns (parse bucket from path, dispatch by method, writeJSON)

**Non-Goals:**
- Actual lifecycle rule enforcement (storage class transitions, object expiration)
- Actual CORS header enforcement (only config storage, no response header injection)
- Changing the existing bucket-level PATCH behavior (keep lifecycle/CORS in BucketUpdateAttrs too)

## Decisions

### Lifecycle and CORS: New Backend interface methods
**Decision**: Add dedicated methods `GetLifecycle`, `SetLifecycle`, `DeleteLifecycle`, `GetCORS`, `SetCORS` to the Backend interface.

**Rationale**: While the data could be accessed via `GetBucket`, dedicated methods align with the REST sub-resource pattern and make the handler code cleaner. The implementation simply delegates to bucket field access — no separate storage maps needed.

**Alternatives considered**:
- Use existing GetBucket/UpdateBucket: Works but requires handler to marshal/unmarshal the full bucket, adds boilerplate to parse/return only lifecycle or CORS fields.

### Lifecycle JSON format
**Decision**: Return lifecycle as `{"kind": "storage#lifecycle", "rule": [...]}` directly, matching GCS API response format.

### CORS JSON format
**Decision**: Return CORS as `{"kind": "storage#cors", "items": [...]}`, matching GCS API response format (wraps cors rules array).

### Health check: Match spec
**Decision**: Change response from `{"status":"ok"}` to `{"status":"healthy","services":{"gcs":"available"}}`.

**Rationale**: The docker-distribution spec explicitly defines this format. Go's SDK health check may not care, but the Docker HEALTHCHECK and monitoring tools might.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Lifecycle/CORS endpoints might be rarely used | Low effort (thin wrappers around existing bucket fields), no wasted work |
| XML DELETE bucket without objects check | Follow GCS behavior: return 409 Conflict if bucket not empty |

## Open Questions

None.
