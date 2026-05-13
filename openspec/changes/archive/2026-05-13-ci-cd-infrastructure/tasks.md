## 1. GitHub Actions CI Workflow

- [x] 1.1 Create `.github/workflows/ci.yml` with `test`, `lint`, and `docker-build` jobs
- [x] 1.2 Configure Go version matrix (1.22, 1.23) for test job
- [x] 1.3 Add `go test -race ./...` step with coverage report
- [x] 1.4 Add `golangci-lint run ./...` step
- [x] 1.5 Add `docker build` step for build verification

## 2. Python SDK Tests

- [x] 2.1 Create `tests/python/` directory with `requirements.txt` (google-cloud-storage)
- [x] 2.2 Write Python test script: bucket CRUD, object upload/download, list with prefix
- [x] 2.3 Add Python SDK test job to CI (spawn emulator, run script, report)

## 3. Node.js SDK Tests

- [x] 3.1 Create `tests/nodejs/` directory with `package.json` (@google-cloud/storage)
- [x] 3.2 Write Node.js test script: bucket CRUD, object upload/download, list with prefix
- [x] 3.3 Add Node.js SDK test job to CI (spawn emulator, run script, report)

## 4. Benchmarks

- [x] 4.1 Add benchmark tests for bucket creation at scale (N buckets)
- [x] 4.2 Add benchmark tests for object upload/download at scale (N objects)
- [x] 4.3 Add benchmark tests for list objects with prefix/delimiter
- [x] 4.4 Add benchmark job to CI (informational, non-blocking)

## 5. Validation

- [x] 5.1 Push to trigger CI and verify all jobs pass
- [x] 5.2 Verify test job catches a deliberate failure
- [x] 5.3 Verify Docker build job works
