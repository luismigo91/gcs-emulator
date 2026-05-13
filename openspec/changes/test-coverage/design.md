## Context

The emulator has 19 services sharing a single HTTP router. Tests use `httptest.Server` with `router.New(b, psb, smb, ctb, kmb, lb, mb, project)` — 7 positional backend params plus a project string. Adding a new service changes the signature, breaking all existing test factories. This creates friction that discourages testing new services.

The actual HTTP handler code for CRUD services is highly regular: parse path, dispatch by method, delegate to backend, return JSON. This makes these services ideal for table-driven testing.

## Goals / Non-Goals

**Goals:**
- Reduce test boilerplate by replacing positional router params with a struct
- Achieve >80% statement coverage across all service packages
- Add tests for all 14 untested services
- Maintain backward compatibility for existing test code

**Non-Goals:**
- 100% branch coverage (edge case testing)
- Performance/load testing
- Replace real backends with mocks (current pattern works)
- Change service handler code

## Decisions

### RouterConfig struct
**Decision**: Add `RouterConfig` struct with named fields. Keep `New()` as a wrapper that delegates to `NewWithConfig(RouterConfig{...})`.

```go
type RouterConfig struct {
    Backend         backend.Backend
    PubSub          pubsubbackend.PubSubBackend
    SecretManager   secretbackend.SecretManagerBackend
    CloudTasks      tasksbackend.CloudTasksBackend
    KMS             kmsbackend.KMSBackend
    Logging         loggingbackend.LoggingBackend
    Monitoring      monitoringbackend.MonitoringBackend
    DefaultProject  string
}
```

**Rationale**: Named fields are self-documenting, `nil` is the zero value, order doesn't matter. Existing callers continue working via `New()` wrapper.

### Table-driven CRUD: One test, many services
**Decision**: Define a `[]crudTestCase` table with endpoint info, body payloads, and expected JSON paths. One test function iterates over the table.

**Rationale**: 10 services × 4 CRUD operations = 40 test cases from one struct definition. Adding a new service is one line in the table.

### Backend access: Return from test factory
**Decision**: Each test helper returns a `TestHarness` struct containing the `*httptest.Server` and all initialized backends.

**Rationale**: Service-specific tests (Nivel 2) need direct backend access for Pub/Sub pull, DLQ verification, BigQuery query execution. A single struct avoids multiple return values.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Table-driven tests may miss service-specific behavior | Nivel 2 tests cover unique behaviors per service |
| RouterConfig adds indirection | Keep `New()` wrapper for backward compat, existing tests unchanged |
| Test runtime increases significantly | Run CRUD tests in parallel with `t.Parallel()` |
| BigQuery SQL tests may be brittle | Test only documented behaviors (SELECT, WHERE, LIMIT) with simple inputs |
