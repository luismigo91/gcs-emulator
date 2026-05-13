## MODIFIED Requirements

### Requirement: Service Backend Configuration
The router SHALL accept backend configuration via a `RouterConfig` struct with named fields, enabling selective backend initialization without positional parameter ordering constraints.

#### Scenario: Create router with only GCS and Pub/Sub
- **WHEN** `NewWithConfig(RouterConfig{Backend: b, PubSub: psb, DefaultProject: "test"})` is called
- **THEN** the returned handler serves both GCS and Pub/Sub routes

#### Scenario: Create router with no backends
- **WHEN** `NewWithConfig(RouterConfig{DefaultProject: "test"})` is called
- **THEN** health, auth, and admin endpoints still work

#### Scenario: Backward compatibility
- **WHEN** `New(b, psb, smb, ctb, kmb, lb, mb, "test")` is called
- **THEN** it delegates to `NewWithConfig` with the same result

### Requirement: Test Harness
The project SHALL provide a reusable test harness that creates an `httptest.Server` with all requested backends and returns both the server and initialized backends for inspection.

#### Scenario: Test harness with specific backends
- **WHEN** `newEmulator(t, RouterConfig{PubSub: psb, DefaultProject: "test"})` is called
- **THEN** a test server is created with Pub/Sub routes available and the PubSub backend is accessible for post-request inspection
