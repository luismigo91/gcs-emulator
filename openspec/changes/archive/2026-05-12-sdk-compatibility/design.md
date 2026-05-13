## Context

We have a working GCS emulator with basic CRUD for buckets and objects, but it's missing critical features that official Google Cloud SDKs depend on. fake-gcs-server (1.4k stars, BSD-2) is the reference implementation. We need to close the gap to be a viable alternative.

**Constraints:**
- Must maintain backward compatibility with existing API surface
- Keep the single-module Go structure
- Storage backends must support new features (generations, chunks, checksums)
- Target: Go SDK + Python SDK + Node.js SDK core operations working

## Goals / Non-Goals

**Goals:**
- List objects with commonPrefixes (delimiter support)
- HTTP 206 Range requests
- int64 generations with proper versioning
- Multipart upload parsing
- Resumable upload sessions with persistent state
- CRC32C and MD5 checksum calculation
- Full response headers (X-Goog-*, ETag, etc.)
- Compose objects (up to 32 sources)
- Complete object metadata fields
- Conditional requests (ifGenerationMatch, etc.)

**Non-Goals:**
- ACLs (bucketAccessControls, objectAccessControls) — defer to next phase
- Full IAM policy enforcement — stub is sufficient for now
- Notifications/PubSub events — defer
- HTTPS/TLS support — defer
- gRPC support — explicitly out of scope
- Folders API — defer
- Managed Folders — defer

## Decisions

### 1. Generations: int64 (nanosecond timestamp)

**Decision**: Use `int64` for generation, generated from `time.Now().UnixNano()`.

**Rationale**: This matches the real GCS API exactly. The Go SDK parses generations as int64. fake-gcs-server uses int64. Our current string-based approach breaks SDK parsing.

**Impact**: Backend interface changes, model changes, all generation comparisons switch to numeric.

### 2. Resumable Upload: In-Memory Session Store

**Decision**: Store upload sessions in a `sync.Map` with chunk data accumulated in memory. Sessions expire after 7 days (matching GCS behavior).

**Rationale**: Simple, fast, matches fake-gcs-server approach. For the MVP, in-memory is sufficient. Persistent sessions could be added later.

**Alternatives considered:**
- File-based chunk storage: More durable but slower, adds filesystem complexity
- Backend-stored chunks: Would require backend interface changes for partial writes

### 3. Checksums: Calculate on Upload

**Decision**: Calculate CRC32C and MD5 when object content is received. Store on the object model. Return in response headers.

**Rationale**: SDKs verify checksums on download. Must be accurate. CRC32C uses the standard Google polynomial (not standard CRC32).

### 4. Conditional Requests: Backend-Level Validation

**Decision**: Pass conditions to the backend interface. Backend validates before performing operations.

**Rationale**: Keeps handler logic clean. Backend has access to current object state for comparison. Matches fake-gcs-server's `backend.Conditions` pattern.

```go
type Conditions struct {
    DoesNotExist        bool
    GenerationMatch     int64
    GenerationNotMatch  int64
    MetagenerationMatch int64
    MetagenerationNotMatch int64
}
```

### 5. CommonPrefixes: Computed at List Time

**Decision**: Compute prefixes during ListObjects by analyzing object names against prefix + delimiter. Don't store separately.

**Rationale**: Matches real GCS behavior. Computation is O(n) over objects in bucket, which is fine for emulator scale. Matches fake-gcs-server approach.

### 6. Multipart Upload: Standard Library `mime/multipart`

**Decision**: Use Go's `mime/multipart` to parse multipart/form-data uploads. Extract JSON metadata from first part, content from second part.

**Rationale**: Standard library, no dependencies. Matches the real GCS multipart format.

### 7. Response Headers: Handler-Level Injection

**Decision**: Add response headers in the handler layer after backend operations complete. Use a helper function to build the standard header set.

**Rationale**: Keeps backend focused on storage. Headers are an API concern, not a storage concern.

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| **CRC32C algorithm mismatch** | High — SDKs reject wrong checksums | Use `github.com/google/cel-go/common/runes` or implement the standard polynomial |
| **Resumable upload memory usage** | Medium — large files consume RAM | Document limitation; add configurable max session size |
| **Generation collisions at nanosecond** | Low — unlikely but possible | Add counter fallback if timestamp repeats |
| **Backend interface breaking changes** | Medium — existing code needs updates | Update all backend implementations together |
| **Multipart parsing edge cases** | Low — standard format, well-defined | Test with SDK-generated requests |
