## 1. Binary Content Persistence

- [x] 1.1 Add blob file write in MemoryBackend.CreateObject (write on mutation)
- [x] 1.2 Add blob file delete in MemoryBackend.DeleteObject
- [x] 1.3 Add blob file delete in MemoryBackend.CreateObject when overwriting without versioning
- [x] 1.4 Add blob read fallback in GetObjectContent for disk modes
- [x] 1.5 Add blob cleanup in MemoryBackend (new CleanupBlobs method for unused blobs)

## 2. Persistent Backend — Extended Save/Load

- [x] 2.1 Extend persistentData struct with IamPolicies and Notifications fields
- [x] 2.2 Update save() to serialize IAM policies and notifications to data.json
- [x] 2.3 Update load() to restore IAM policies and notifications from data.json
- [x] 2.4 Override GetObjectContent to try blob files before zeroed buffers
- [x] 2.5 Add backward-compatible loading (handle missing fields in old data.json)

## 3. Hybrid Backend — Extended Save/Load

- [x] 3.1 Apply same persistentData struct changes (shared with persistent)
- [x] 3.2 Update save() to serialize IAM policies and notifications
- [x] 3.3 Update load() to restore IAM policies and notifications
- [x] 3.4 Mark IAM/notification mutations as dirty (trigger async flush)

## 4. WAL Backend — Extended Operations

- [x] 4.1 Add WAL operation types: set_iam_policy, create_notification, delete_notification
- [x] 4.2 Override SetBucketIAMPolicy to append WAL entry before applying
- [x] 4.3 Override CreateNotification to append WAL entry before applying
- [x] 4.4 Override DeleteNotification to append WAL entry before applying
- [x] 4.5 Implement IAM/notification operation replay in applyOperation()
- [x] 4.6 Extend compaction snapshot to include IAM policies and notifications

## 5. Auth Token Endpoint

- [x] 5.1 Create `internal/auth/handler.go` with OAuth2 token endpoint handler
- [x] 5.2 Handle both form-encoded and JSON Content-Type bodies
- [x] 5.3 Return well-formed token response: access_token, token_type, expires_in, scope
- [x] 5.4 Register `/oauth2/v4/token` route in router.go

## 6. Testing

- [x] 6.1 Test blob write and read roundtrip in persistent mode
- [x] 6.2 Test blob cleanup on object delete in persistent mode
- [x] 6.3 Test IAM policy survives restart in persistent mode
- [x] 6.4 Test notification survives restart in persistent mode
- [x] 6.5 Test hybrid flush includes IAM and notifications
- [x] 6.6 Test WAL replay of IAM and notification operations
- [x] 6.7 Test WAL compaction includes IAM and notifications
- [x] 6.8 Test backward-compatible loading of old data.json format
- [x] 6.9 Test auth token endpoint returns valid response
- [x] 6.10 Test auth token endpoint accepts both form and JSON body
- [x] 6.11 Run full test suite with `go test -race ./...`
