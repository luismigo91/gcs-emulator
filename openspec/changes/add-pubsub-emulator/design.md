## Context

The emulator currently serves only GCS on a single HTTP server. The router, middleware, backend factory, and config system are all service-agnostic — they don't hardcode "GCS" anywhere except in route registration. Adding a second service is an architectural validation of the existing design.

The Pub/Sub REST API follows the pattern `https://pubsub.googleapis.com/v1/projects/{project}/topics/{topic}`. All resources are project-scoped, matching our existing `projectMiddleware`.

## Goals / Non-Goals

**Goals:**
- Full Pub/Sub REST API (topics, subscriptions, publish, pull, ack, schemas, snapshots)
- GCS object events trigger Pub/Sub message delivery to configured notification topics
- Same storage backends (memory, persistent, hybrid, WAL) for message persistence
- Single binary, single port, same Docker image
- Go SDK compatibility via `pubsub.NewClient(ctx, project, option.WithEndpoint(...))`

**Non-Goals:**
- gRPC API (REST only, matching existing GCS pattern)
- Push subscriptions (pull-only for emulator scope)
- Exactly-once delivery semantics
- BigQuery subscriptions
- Cloud Storage subscriptions (cross-service subscription types)
- Real authentication or IAM enforcement for Pub/Sub
- Ordered delivery or message ordering keys enforcement
- Regional/global topic replication

## Decisions

### Architecture: In-process, single port
**Decision**: Pub/Sub runs in the same Go process as GCS, sharing the `net/http` ServeMux. No separate binary, no separate port.

**Rationale**: The project's value prop is a single 15MB binary. Multi-port or sidecar adds operational complexity. The `http.ServeMux` already supports prefix-based routing — Pub/Sub at `/v1/projects/`, GCS at `/storage/v1/`.

**Alternatives considered**:
- Separate binary: Breaks the "one binary" promise, harder to coordinate
- Separate port: Same binary, different listener. More config, more Docker complexity
- gRPC-only: Some SDKs prefer gRPC, but REST-first matches existing GCS pattern

### Backend: Shared interface pattern
**Decision**: Define `PubSubBackend` interface with `Topic`, `Subscription`, `Message` storage. Implement 4 backends following the same factory pattern as GCS.

**Rationale**: Reuses all infrastructure (blob persistence, IAM/notification persistence patterns). Zero new architectural concepts. Messages are just byte blobs with metadata — identical to GCS objects.

### GCS→Pub/Sub bridge: Direct call, no network hop
**Decision**: When GCS creates/deletes an object in a bucket with notifications, call `PubSubBackend.Publish()` directly in-process. No HTTP call, no queue.

**Rationale**: The emulator runs everything in-process. Making an HTTP call to itself is wasteful and introduces failure modes. The bridge is a pure Go interface call.

### Message storage: Append-only log per subscription
**Decision**: Each subscription maintains an ordered list of `Message` entries (message_id, data, attributes, publish_time, ack_id). Pull returns up to `maxMessages` unacknowledged messages. Ack removes the message.

**Rationale**: Simple, matches Pub/Sub semantics closely enough for testing. Re-delivery after ack deadline expiry is a nice-to-have (not in v1).

### Schemas: Storage-only validation
**Decision**: Store schema definitions (Avro/Protobuf) and validate message data against schema on publish if schema is attached to topic. `encoding` field on schema resource controls validation type.

**Rationale**: Schema validation is important for testing data contracts. No external Avro/Protobuf compiler — just syntax-check mode, or accept-without-validation mode for simplicity.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Memory usage grows with unacknowledged messages | Configurable max unacknowledged per subscription, oldest expire first |
| No push subscriptions limits some integration tests | Document as non-goal; pull subscriptions cover 90% of use cases |
| Schema validation could break if message format is wrong | Make validation optional (configurable, default: warn but accept) |
| Message ordering not enforced | Accept limitation v1; ordering keys stored but not enforced |

## Open Questions

- Should we support Pub/Sub Lite (regional, lower latency, no schema)?
  → Answer: No. Lite is a different API surface. Standard Pub/Sub is the right level.
- Should we support streaming pull (long-lived gRPC stream)?
  → Answer: Not in v1. Sync pull via REST covers test scenarios.
