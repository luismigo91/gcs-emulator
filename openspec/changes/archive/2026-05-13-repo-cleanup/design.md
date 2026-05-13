## Context

The repo was scaffolded with a future Go version flag and lacks standard OSS artifacts (LICENSE, linter config). Two OpenSpec changes are in inconsistent states: one completed but unarchived, one with stale tasks.

## Goals / Non-Goals

**Goals:**
- Add missing standard repo files
- Fix broken Go module version
- Bring OpenSpec artifacts in sync with code

**Non-Goals:**
- No code changes
- No spec changes
- No test additions

## Decisions

### Go version: 1.22.0
**Decision**: Set `go 1.22.0` in go.mod.
**Rationale**: Go 1.22 is widely adopted, still supported, and the code doesn't use any post-1.22 language features (no range-over-func, no new iterators). The current `1.26.3` is impossible.

### Linter: Standard golangci-lint config
**Decision**: Use a minimal `.golangci.yml` enabling `errcheck`, `gosimple`, `govet`, `ineffassign`, `staticcheck`, `unused`.
**Rationale**: Standard Go linters, no exotic rules. Matches `make lint` expectation in CONTRIBUTING.md.

### OpenSpec sync: Manual task update
**Decision**: Mark IAM (2.x) and Notifications (3.x) task groups as `[x]` in gcs-emulator-complete/tasks.md. Archive integration-tests-core.
**Rationale**: These are already implemented in code. The tasks file is documentation debt, not code debt.
