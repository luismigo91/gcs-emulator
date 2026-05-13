## ADDED Requirements

### Requirement: Automated test execution
The project SHALL run `go test -race ./...` on every push and pull request via GitHub Actions.

#### Scenario: Tests run on push
- **WHEN** a commit is pushed to any branch
- **THEN** GitHub Actions executes `go test -race ./...` and reports pass/fail

#### Scenario: Tests run on PR
- **WHEN** a pull request is opened against main
- **THEN** GitHub Actions executes the test suite and blocks merge on failure

### Requirement: Automated linting
The project SHALL run `golangci-lint` on every push and pull request.

#### Scenario: Lint runs on push
- **WHEN** a commit is pushed
- **THEN** `golangci-lint run ./...` executes and reports any violations

### Requirement: Docker build verification
The project SHALL verify the Docker image builds successfully on every push and pull request.

#### Scenario: Docker build succeeds in CI
- **WHEN** the CI pipeline runs
- **THEN** `docker build -t gcs-emulator .` completes without errors

### Requirement: Go version matrix
The CI pipeline SHALL test against at least two Go versions to catch version-specific issues.

#### Scenario: Multi-version test matrix
- **WHEN** the CI pipeline runs
- **THEN** tests execute on Go 1.22 and Go 1.23
