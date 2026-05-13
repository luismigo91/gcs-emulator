## Why

The path parsing in `ExtractPathParams()` uses string prefix matching and manual splitting to extract `bucket`, `object`, and other parameters from URLs. This approach is fragile and has already broken twice when the SDK sent different URL formats than expected (e.g., `/rewriteTo/b/{bucket}/o/{object}` vs `/rewriteTo/{bucket}/o/{object}`). Each break required patching both the parser and the handler, creating a maintenance burden and risk of regressions.

## What Changes

- Replace `ExtractPathParams()` with a pattern-based router that matches URL templates
- Move copyTo and rewriteTo routing from handler string matching to explicit route registration
- Split `BucketHandler` into dedicated route handlers for each endpoint pattern
- Keep the existing `http.ServeMux` approach (Go 1.22+ supports pattern matching) but use it more effectively

## Capabilities

### New Capabilities
- `path-routing`: Pattern-based URL routing with explicit endpoint registration

### Modified Capabilities
- `gcs-api`: Routing is an implementation detail, not a spec change. No requirement modifications needed.

## Impact

- `internal/util/params.go`: `ExtractPathParams()` replaced or removed
- `internal/router/router.go`: New route registration with patterns
- `internal/api/handler.go`: Simplified handlers (no more path parsing logic)
- All existing tests should continue to pass (behavior unchanged)
