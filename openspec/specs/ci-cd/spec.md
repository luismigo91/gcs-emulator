# ci-cd

## Requirements

### Requirement: GitHub Actions CI
The emulator SHALL have automated CI that runs tests, linting, and Docker build verification.

#### Scenario: Tests run on push
- **WHEN** a commit is pushed to any branch
- **THEN** go test -race ./... executes in CI

### Requirement: Multi-version Go Matrix
CI SHALL test against Go 1.22 and 1.23.

#### Scenario: Tests pass on both versions
- **WHEN** CI runs
- **THEN** test and build jobs execute on both Go versions

### Requirement: Docker Build Verification
CI SHALL verify the Docker image builds successfully.

#### Scenario: Docker build succeeds
- **WHEN** CI docker-build job runs
- **THEN** docker build completes without errors

### Requirement: Benchmark Execution
CI SHALL run benchmarks without blocking the pipeline.

#### Scenario: Benchmarks run as informational
- **WHEN** CI runs
- **THEN** go test -bench executes and results are captured as artifacts
