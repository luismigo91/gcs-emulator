## 1. Pub/Sub Models and Backend

- [x] 1.1 Create `internal/pubsub/model/` — Topic, Subscription, Message, Schema types
- [x] 1.2 Create `internal/pubsub/backend/backend.go` — PubSubBackend interface
- [x] 1.3 Implement MemoryPubSubBackend
- [ ] 1.4 Implement PersistentPubSubBackend
- [ ] 1.5 Implement HybridPubSubBackend
- [ ] 1.6 Implement WALPubSubBackend
- [x] 1.7 Create `internal/pubsub/backend/factory.go`

## 2. Pub/Sub API Handlers

- [x] 2.1 Create topic handler (CRUD + Publish)
- [x] 2.2 Create subscription handler (CRUD + Pull + Ack + ModifyAckDeadline)
- [x] 2.3 Create schema handler (CRUD + Validate + Commit)
- [x] 2.4 writeJSON/writeError helpers

## 3. Multi-Service Routing

- [x] 3.1 Refactor `internal/router/` to accept PubSubBackend
- [x] 3.2 Register Pub/Sub routes
- [x] 3.3 Update health check for both gcs + pubsub
- [x] 3.4 Update `cmd/server/main.go` to start Pub/Sub

## 4. GCS → Pub/Sub Bridge

- [x] 4.1 triggerNotification on object create
- [x] 4.2 Filter by event type and object name prefix
- [x] 4.3 Build message attributes (bucketId, objectId, eventType)
- [x] 4.4 Log failures without blocking GCS operation

## 5. Config and Emulator Package

- [x] 5.1 Extend config for Pub/Sub (GCP_EMULATOR_PUBSUB_MODE env var)
- [x] 5.2 Update pkg/emulator for PubSubConfig
- [x] 5.3 Update Server.Stop() to shutdown Pub/Sub

## 6. Testing

- [x] 6.1 Test topic CRUD via HTTP API
- [x] 6.2 Test subscription CRUD via HTTP API
- [x] 6.3 Test publish → pull → ack flow
- [x] 6.4 Test modify ack deadline
- [x] 6.5 Test schema CRUD
- [x] 6.6 Test GCS object create triggers Pub/Sub
- [x] 6.7 Test GCS object delete triggers Pub/Sub (OBJECT_DELETE wired)
- [ ] 6.8 Test multiple subscriptions receive same message
- [ ] 6.9 Test no-op when no notification config
- [ ] 6.10 Test project isolation for Pub/Sub resources
- [ ] 6.11 Test persistent/hybrid/WAL backends for Pub/Sub
- [x] 6.12 Run `go test -race ./...` and confirm all pass

## 7. SDK and Docker

- [ ] 7.1 Add Pub/Sub Go SDK test
- [ ] 7.2 Add Pub/Sub Python SDK test
- [ ] 7.3 Add Pub/Sub Node.js SDK test
- [ ] 7.4 Update Docker health check
- [ ] 7.5 Update CI workflow
