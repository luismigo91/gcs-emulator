## Why

The repository has several hygiene issues that undermine professionalism: missing LICENSE file despite README claiming MIT, an impossible Go version in go.mod (`1.26.3` doesn't exist), no linter configuration, and stale OpenSpec artifacts that are out of sync with reality.

## What Changes

- Add `LICENSE` file (MIT) matching README
- Fix `go.mod`: change `go 1.26.3` to `go 1.22.0` (stable, widely supported)
- Add `.golangci.yml` configuration for `make lint`
- Archive `integration-tests-core` change (all tasks completed)
- Sync `gcs-emulator-complete` tasks.md to reflect actual implementation status (IAM and notifications are done, marked as pending)

## Capabilities

No functional capabilities are added or modified. This is a repo maintenance change.

## Impact

- Root files: `LICENSE` (new), `go.mod` (modified), `.golangci.yml` (new)
- `openspec/changes/integration-tests-core/` → archived
- `openspec/changes/gcs-emulator-complete/tasks.md` → synced with reality
