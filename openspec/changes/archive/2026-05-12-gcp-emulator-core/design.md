## Context

This is a greenfield project — a GCP emulator in Go, inspired by Floci's architecture for AWS. The existing landscape includes `fake-gcs-server` (GCS-only, BSD-2, Go) and LocalStack (AWS, Python, archived). The goal is to build a unified GCP emulator starting with GCS, designed from day one for multi-service expansion.

**Constraints:**
- Must be wire-compatible with official Google Cloud SDKs (Go, Python, Node.js, Java)
- GCP uses both REST (JSON + XML) and gRPC protocols — GCS is REST-only, but future services will need gRPC
- Go ecosystem has excellent gRPC/protobuf support via `google.golang.org/grpc`
- Target: ~15-20MB binary, <100ms startup, <30MB idle RAM

## Goals / Non-Goals

**Goals:**
- GCS API emulation compatible with all major SDKs (JSON API v1 + XML API)
- Four storage modes: memory, persistent, hybrid, WAL
- Multi-project resource isolation via `GOOGLE_CLOUD_PROJECT`
- Docker-first distribution with environment variable configuration
- Plugin-based service architecture for future expansion
- Go library API for programmatic use in tests
- MIT license, free forever

**Non-Goals:**
- Real gRPC emulation (MVP is REST-only; gRPC comes with Pub/Sub/Firestore)
- Signed URL validation (accept any signed URL, like fake-gcs-server)
- Real GCP credential validation (accept any credentials)
- Web UI dashboard (future phase)
- Cloud Functions / Cloud Run / Cloud SQL emulation (future phases with Docker integration)

## Decisions

### 1. HTTP Router: Standard Library `net/http` + `http.ServeMux`

**Decision**: Use Go's standard library router instead of third-party routers like Gorilla Mux or Chi.

**Rationale**: Go 1.22+ `http.ServeMux` supports path patterns with wildcards (`/storage/v1/b/{bucket}/o/{object}`), method matching, and has zero dependencies. For a performance-critical emulator, avoiding dependency overhead matters.

**Alternatives considered:**
- Gorilla Mux: Mature but unmaintained, adds dependency
- Chi: Lightweight but unnecessary complexity for our use case
- FastHTTP: Faster but incompatible with standard `http.Handler` ecosystem

### 2. Storage Backend: In-Memory Maps with Pluggable Persistence

**Decision**: Primary storage is `map[string]*Object` and `map[string]*Bucket` in RAM. Persistence modes serialize to disk (JSON format) on different schedules.

**Rationale**: GCS objects are typically small-to-medium files. In-memory maps give O(1) lookups, simple concurrency with `sync.RWMutex`, and trivial serialization. The WAL mode uses append-only JSON lines for crash recovery.

**Alternatives considered:**
- SQLite: Overkill for key-value access patterns, adds CGo dependency
- BoltDB/Badger: Good but adds complexity; JSON files are human-readable and debuggable
- Filesystem-only (fake-gcs-server approach): Slow for listing/metadata operations

### 3. JSON Serialization: Standard `encoding/json`

**Decision**: Use Go's standard library JSON encoder/decoder.

**Rationale**: GCS API responses are well-defined JSON structures. Standard library is fast enough and avoids dependency bloat. For the hot path (object upload/download), we stream bytes directly without JSON overhead.

### 4. Concurrency Model: Per-Backend `sync.RWMutex`

**Decision**: Each storage backend instance has its own `sync.RWMutex`. No global lock.

**Rationale**: Fine-grained locking allows concurrent operations on different buckets/objects. For the MVP scope, this is sufficient. Future optimization could use sharded locks if needed.

### 5. Configuration: Environment Variables with `GCP_EMULATOR_` Prefix

**Decision**: All configuration via environment variables, not config files or CLI flags (flags can set env vars at startup).

**Rationale**: Docker-first design means env vars are the natural configuration mechanism. Matches Floci's `FLOCI_*` and LocalStack's `LOCALSTACK_*` patterns.

**Key variables:**
- `GCP_EMULATOR_PORT` (default: `9090`)
- `GCP_EMULATOR_DEFAULT_PROJECT` (default: `test-project`)
- `GCP_EMULATOR_STORAGE_MODE` (default: `memory`)
- `GCP_EMULATOR_STORAGE_PATH` (default: `./data`)

### 6. Project Isolation: Access Key / Project Header Resolution

**Decision**: Project ID resolved from (in priority order):
1. `X-Goog-User-Project` header
2. `GOOGLE_CLOUD_PROJECT` environment variable on the client request
3. `GCP_EMULATOR_DEFAULT_PROJECT` server configuration

**Rationale**: Matches how the real GCP SDK sends project context. The `X-Goog-User-Project` header is the standard GCP mechanism.

### 7. Error Responses: GCP JSON Error Format

**Decision**: All errors follow the GCP JSON error format:
```json
{
  "error": {
    "code": 404,
    "message": "Not Found",
    "errors": [...]
  }
}
```

**Rationale**: SDKs parse this format. Must match for compatibility.

### 8. Object Storage: Bytes in Memory, Not on Disk (for memory mode)

**Decision**: Object content stored as `[]byte` in memory, not written to filesystem even in memory mode.

**Rationale**: Simpler, faster, and avoids filesystem permission issues. Persistent/hybrid/WAL modes handle disk serialization.

### 9. Go Module Structure: Single Module with Internal Packages

**Decision**: Single `go.mod` at the root. `internal/` for private packages, `pkg/` for the public library API.

**Rationale**: Simpler than multi-module. The `pkg/emulator` package provides the public API for Go library usage.

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| **SDK compatibility gaps** | High — core value proposition | Extensive compatibility test suite across Go, Python, Node.js, Java SDKs |
| **In-memory limits for large objects** | Medium — multi-GB objects will consume RAM | Document limitations; persistent mode can stream to disk for large objects |
| **GCS API surface is large** | Medium — 50+ endpoints | Prioritize most-used endpoints first; use SDK test suites to identify gaps |
| **Go 1.22+ ServeMux pattern limitations** | Low — edge cases in routing | Can swap to Chi if needed; standard library covers 95% of patterns |
| **JSON serialization performance** | Low — metadata operations only | Hot path (upload/download) bypasses JSON; metadata ops are infrequent |
| **No gRPC in MVP** | Low — GCS is REST-only | gRPC infrastructure planned for Pub/Sub phase; design supports it |
| **Concurrency bugs with RWMutex** | Medium — data corruption | Comprehensive concurrent access tests; consider `sync.Map` for hot paths |
