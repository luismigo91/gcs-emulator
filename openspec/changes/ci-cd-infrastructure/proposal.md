## Why

The emulator lacks automated CI/CD, making it impossible to catch regressions on push or PR. SDK compatibility tests for Python and Node.js are missing, leaving unknown gaps for those language ecosystems. Performance benchmarks don't exist, so there's no way to measure impact of changes on throughput or latency.

## What Changes

### CI/CD Pipeline
- GitHub Actions workflow: `go test -race`, `golangci-lint`, `docker build` on push and PR
- Multi-version Go matrix (1.22, 1.23)

### SDK Compatibility Tests
- Python: Test script using `google-cloud-storage` against running emulator
- Node.js: Test script using `@google-cloud/storage` against running emulator

### Performance Benchmarks
- Go benchmark tests for bucket creation, object upload/download at scale (100, 1000 objects)
- Benchmark results stored as GitHub Action artifacts

## Capabilities

### New Capabilities
- `ci-cd-pipeline`: Automated test, lint, and build on push/PR via GitHub Actions
- `sdk-python-tests`: Python SDK compatibility test suite
- `sdk-nodejs-tests`: Node.js SDK compatibility test suite
- `benchmarks`: Performance benchmarking infrastructure for bucket and object operations

## Impact

- `.github/workflows/ci.yml`: GitHub Actions workflow (new)
- `tests/python/`: Python SDK compatibility tests (new)
- `tests/nodejs/`: Node.js SDK compatibility tests (new)
- `internal/backend/benchmark_test.go`: Already exists, may extend
