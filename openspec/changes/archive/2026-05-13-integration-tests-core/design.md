## Context

We have a working GCS emulator with ~55% API coverage. The SDK compatibility change added critical features (generations, checksums, range requests, conditional requests, etc.) but we haven't validated them against real SDKs. The remaining work falls into three categories:

1. **Integration tests** — HTTP-level tests that verify request/response shapes
2. **SDK compatibility tests** — End-to-end tests using official SDKs
3. **Missing core features** — HEAD requests, data preloading, concurrent access validation

**Constraints:**
- Tests must run without real GCP credentials
- SDK tests need the emulator running as a subprocess or test server
- Data preloading must work both in Docker (volume mount) and as a Go library

## Goals / Non-Goals

**Goals:**
- Integration tests for all bucket API endpoints (CRUD, IAM, lifecycle)
- Integration tests for all object API endpoints (CRUD, compose, copy, rewrite)
- Integration tests for resumable uploads
- Integration tests for XML API endpoints
- Integration tests for versioning and lifecycle
- Concurrent access tests with race detection
- Graceful shutdown tests
- HEAD request support for bucket and object existence checks
- Data preloading from directory structure
- Go SDK compatibility test suite

**Non-Goals:**
- Python SDK compatibility tests (defer — requires Python environment setup)
- Node.js SDK compatibility tests (defer — requires Node.js environment setup)
- IAM/notifications integration tests (defer — IAM is stub-only)
- Architecture diagram (separate effort)

## Decisions

### 1. Integration Tests: httptest.Server

**Decision**: Use Go's `httptest.NewServer` to spin up the emulator in-process for each test.

**Rationale**: Fast, no network flakiness, easy to inspect responses. Matches fake-gcs-server's approach.

### 2. SDK Compatibility Tests: Go Only (MVP)

**Decision**: Start with Go SDK compatibility tests. Python and Node.js tests require additional CI setup.

**Rationale**: Go SDK is the primary target (we're building in Go). Python/Node.js can be added later once the Go tests pass consistently.

### 3. HEAD Requests: Reuse GET Handlers

**Decision**: HEAD requests reuse the same handler logic as GET, but skip writing the response body.

**Rationale**: Matches HTTP spec. GCS HEAD returns the same headers as GET but no body. Simple to implement.

### 4. Data Preloading: Directory Scan on Startup

**Decision**: On startup, scan the seed directory for `{bucket-name}/{object-path}` structure. Create buckets and objects from files found.

**Rationale**: Matches fake-gcs-server exactly. Simple, intuitive, works with Docker volume mounts.

### 5. Concurrent Access: go test -race

**Decision**: Run all tests with `-race` flag. Fix any data races found.

**Rationale**: Go's race detector is the gold standard. If tests pass with `-race`, concurrent access is safe.

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| **SDK version drift** | Medium — SDK APIs change over time | Pin SDK versions in go.mod, test against latest |
| **Large test suite slows CI** | Low — tests run in-process | Parallelize tests, use t.Parallel() |
| **Data preload memory usage** | Medium — large seed directories consume RAM | Document limitation, streaming mode for large files |
| **Race conditions in backends** | High — data corruption | Run all tests with -race, fix before merge |
