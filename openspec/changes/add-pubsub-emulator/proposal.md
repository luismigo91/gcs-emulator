## Why

GCS bucket notifications are currently storage-only — configurations are saved but never delivered. Adding Pub/Sub emulation unlocks end-to-end event-driven workflows (GCS → Pub/Sub → Cloud Functions) in local development. Beyond GCS integration, Pub/Sub is the most-used GCP service that lacks a lightweight Go emulator — the official emulator requires Java and the gcloud SDK (~500MB).

This transforms the repo from a single-service emulator into a **multi-service GCP emulator** with zero extra infrastructure cost — reusing the same binary, storage backends, project isolation, auth, metrics, and Docker image.

## What Changes

### Pub/Sub Core API
- **Topics**: create, get, list, delete, publish
- **Subscriptions**: create, get, list, delete, pull, acknowledge, modifyAckDeadline
- **Schemas**: create, get, list, delete, validate, commit
- **Snapshots/Seek**: create, get, list, delete, seek

### GCS + Pub/Sub Integration
- When a GCS object event occurs (OBJECT_FINALIZE, OBJECT_DELETE, etc.), the emulator publishes a Pub/Sub message to any configured notification topics
- Subscribers can pull/ack messages, enabling end-to-end testing

### Multi-Service Architecture
- Each service (GCS, Pub/Sub) runs in-process on the same port
- Services share: storage backend, project middleware, auth, metrics, config
- New `pkg/emulator` API extended to start all services

## Capabilities

### New Capabilities
- `pubsub-topic-api`: Topic management (CRUD + publish)
- `pubsub-subscription-api`: Subscription management with pull/ack/deadline
- `pubsub-schema-api`: Schema management for Avro/Protobuf
- `gcs-pubsub-integration`: GCS object events trigger real Pub/Sub message delivery
- `multi-service-architecture`: Runtime that serves multiple GCP services from one binary

### Modified Capabilities
- `bucket-notifications`: Notifications now deliver messages to Pub/Sub instead of being storage-only
- `docker-distribution`: Docker image now runs both GCS and Pub/Sub

## Impact

- `internal/pubsub/`: New package (models, API handlers, Pub/Sub backend)
- `internal/backend/`: Pub/Sub backend interface + 4 implementations (extends MemoryBackend pattern)
- `internal/api/`: GCS handler modified to publish Pub/Sub messages on object events
- `internal/router/`: Pub/Sub routes registered alongside GCS routes
- `internal/config/`: Per-service config overrides for Pub/Sub mode/path
- `pkg/emulator/`: Multi-service server start/stop
- `cmd/server/`: Main now starts both services
- `tests/pubsub/`: Integration tests
- `Makefile`: New test target
