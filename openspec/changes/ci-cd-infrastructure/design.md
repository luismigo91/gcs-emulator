## Context

The project has a Makefile with `build`, `test`, `lint`, and `docker-build` targets but no automated execution. SDK tests exist for Go (`pkg/emulator/sdk_test.go`) but not Python or Node.js. A benchmark test file exists (`internal/api/benchmark_test.go`) but is limited.

## Goals / Non-Goals

**Goals:**
- CI pipeline running on every push and PR to main
- Go test with race detection, linter, and Docker build verification
- Python and Node.js SDK smoke tests (bucket CRUD, object upload/download)
- Go benchmarks for core operations

**Non-Goals:**
- Cross-platform CI matrix (Linux only, matching Docker target)
- CD (no auto-publish to container registries)
- Comprehensive SDK test coverage (smoke tests only, document gaps)
- Benchmark regression gates (collect data, don't block on it yet)

## Decisions

### CI: Single workflow with jobs
**Decision**: One `ci.yml` workflow with parallel jobs: `test`, `lint`, and `docker-build`.

**Rationale**: Simple, fast. Go test and lint can't run meaningfully before the other. Docker build verifies the image builds but doesn't test it.

### SDK tests: Dockerized
**Decision**: Run Python and Node.js tests inside Docker containers in CI to avoid requiring both language toolchains on the CI runner.

**Rationale**: Keeps the CI runner simple (just Go + Docker). Tests spin up the emulator in a container, run SDK scripts against it, report results.

**Alternatives considered**:
- Install Python/Node.js on CI runner: More complex setup, version management headaches.
- Separate test repos: Fragments the project, harder to track compatibility.

### Benchmarks: Go native
**Decision**: Use Go's `testing.B` benchmark framework, run via `go test -bench=. ./...` in CI.

**Rationale**: No external tools needed. Results are comparable across runs.

### CI trigger: Push to any branch + PR to main
**Decision**: Run on `push` (all branches) and `pull_request` (targeting main).

**Rationale**: Catch issues early on feature branches, verify PRs before merge.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| SDK tests flaky due to network timing | Retry logic in test scripts, generous timeouts |
| Python/Node.js SDKs have heavy dependencies | Use pre-built Docker images with SDKs installed |
| Benchmark variability in CI | Use `-benchtime=5s` for stable results, document that CI benchmarks are approximate |

## Open Questions

None.
