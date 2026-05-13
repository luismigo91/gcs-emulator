## 1. Router Refactoring

- [x] 1.1 Update `router.go` to register explicit route patterns using Go 1.22+ ServeMux wildcards
- [x] 1.2 Replace `ExtractPathParams()` usage with `r.PathValue()` calls in handlers
- [x] 1.3 Register separate routes for copyTo and rewriteTo with both URL format variants
- [x] 1.4 Remove `ExtractPathParams()` from `internal/util/params.go`

## 2. Handler Simplification

- [x] 2.1 Simplify `BucketHandler` to only handle bucket-level operations (no object routing)
- [x] 2.2 Update `ObjectHandler` to use `r.PathValue()` instead of params map
- [x] 2.3 Update `UploadHandler` and `ResumableUploadHandler` to use `r.PathValue()`

## 3. Validation

- [x] 3.1 Run `go test -race ./...` and confirm all tests pass
- [x] 3.2 Manual test: verify copyTo and rewriteTo work with both URL formats via curl
