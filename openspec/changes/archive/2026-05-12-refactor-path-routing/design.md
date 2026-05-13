## Context

The current routing architecture uses `http.ServeMux` with two broad patterns (`/storage/v1/b` and `/storage/v1/b/`) that funnel all requests into `BucketHandler`. The handler then uses `ExtractPathParams()` to manually parse the URL path and determine which sub-handler to call. This creates a two-level routing system where the first level is implicit (ServeMux patterns) and the second level is manual string parsing.

The `copyTo` and `rewriteTo` endpoints have two valid URL formats from different SDK versions, and the parser only handles one format at a time, requiring patches.

## Goals / Non-Goals

**Goals:**
- Eliminate manual path parsing in favor of explicit route registration
- Support both URL formats for copyTo and rewriteTo without string matching hacks
- Keep `http.ServeMux` (no new dependencies)
- Maintain backward compatibility with all existing tests

**Non-Goals:**
- Switch to a third-party router (chi, gorilla/mux, etc.)
- Add gRPC routing support
- Change the handler function signatures

## Decisions

### Use Go 1.22+ ServeMux patterns

**Decision**: Leverage Go 1.22+ `http.ServeMux` support for path patterns with `{name}` wildcards.

**Rationale**: The project already uses Go 1.26.3. Go 1.22+ ServeMux supports wildcard patterns like `/storage/v1/b/{bucket}/o/{object...}` which capture path segments including slashes. This eliminates the need for manual parsing without adding dependencies.

**Alternative considered**: Switch to `chi` or `gorilla/mux`. Rejected because it adds a dependency and the stdlib now supports what we need.

### Register explicit routes for copyTo and rewriteTo

**Decision**: Register separate route patterns for `/storage/v1/b/{bucket}/o/{object}/copyTo/b/{destBucket}/o/{destObject}` and the shorter format.

**Rationale**: Each URL format gets its own handler function. No string parsing needed. The ServeMux matches the most specific pattern first.

### Extract object operations into a dedicated handler group

**Decision**: Create `ObjectHandler` that handles all object-level operations (get, delete, patch, head) registered at `/storage/v1/b/{bucket}/o/{object...}`.

**Rationale**: Currently `BucketHandler` handles both bucket and object operations. Separating them makes the routing clearer and reduces the complexity of the catch-all handler.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Go 1.22+ pattern matching may not handle all edge cases | The existing test suite covers all URL patterns; tests will catch mismatches |
| Refactoring may break subtle URL parsing behavior | Run full test suite after each task; manual curl verification |
| Handler splitting increases file count | Keep handlers in `internal/api/handler.go` but organized by resource |
