## 1. Repo Files

- [x] 1.1 Create `LICENSE` file with MIT license text
- [x] 1.2 Fix `go.mod`: change `go 1.26.3` to `go 1.22.0`
- [x] 1.3 Create `.golangci.yml` with standard linter configuration

## 2. OpenSpec Sync

- [x] 2.1 Archive `integration-tests-core` change (all tasks completed)
- [x] 2.2 Mark IAM task group (2.1-2.5) as `[x]` in `gcs-emulator-complete/tasks.md`
- [x] 2.3 Mark Notification task group (3.1-3.5) as `[x]` in `gcs-emulator-complete/tasks.md`
- [x] 2.4 Mark Metrics middleware tasks (8.1-8.2) as `[x]` in `gcs-emulator-complete/tasks.md`
- [x] 2.5 Mark Docker optimization task (8.4) as `[x]` in `gcs-emulator-complete/tasks.md`

## 3. Validation

- [x] 3.1 Run `make build` to verify go.mod version works
- [x] 3.2 Run `make lint` to verify golangci config works
- [x] 3.3 Run `make test` to confirm no regressions
