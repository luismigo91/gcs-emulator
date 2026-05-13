## ADDED Requirements

### Requirement: Object operation benchmarks
The project SHALL include Go benchmark tests for core object operations to measure performance and detect regressions.

#### Scenario: Benchmark bucket creation
- **WHEN** `go test -bench=BenchmarkBucket -benchtime=5s ./internal/...` is run
- **THEN** benchmark results are produced for bucket creation at scale

#### Scenario: Benchmark object upload
- **WHEN** `go test -bench=BenchmarkObject -benchtime=5s ./internal/...` is run
- **THEN** benchmark results are produced for object upload and download at scale (100+ objects)

#### Scenario: Benchmark list objects
- **WHEN** `go test -bench=BenchmarkList -benchtime=5s ./internal/...` is run
- **THEN** benchmark results are produced for listing objects with prefix/delimiter

### Requirement: CI benchmark execution
Benchmarks SHALL run in CI without blocking the pipeline (informational only).

#### Scenario: Benchmarks run in CI
- **WHEN** the CI pipeline executes
- **THEN** benchmarks run and results are captured as job artifacts, but benchmark performance does not block pipeline success
