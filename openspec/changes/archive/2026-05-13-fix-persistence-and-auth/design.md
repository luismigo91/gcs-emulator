## Context

The Persistent, Hybrid, and WAL backends all embed `*MemoryBackend` and override write methods to add disk persistence. However, only `buckets` and `objects` maps are serialized — `content` (binary data), `iamPolicies`, and `notifications` maps are excluded from save/load. This means:

- **Object content**: On reload, `persistent.go:138` allocates zeroed buffers of the correct size (`make([]byte, obj.Size)`) — byte content is lost
- **IAM policies**: Set at runtime via `SetBucketIAMPolicy`, never written to disk, lost on restart
- **Notifications**: CRUD at runtime, never persisted, lost on restart

The `internal/auth/` package has a directory but zero files. The auth-emulation spec (`openspec/specs/auth-emulation/spec.md`) defines requirements for an OAuth2 token endpoint that SDKs like `google-cloud-storage` (Python) and `@google-cloud/storage` (Node.js) use during their credential flow.

## Goals / Non-Goals

**Goals:**
- Persist binary object content alongside metadata in all disk-backed modes
- Persist IAM policies and notification configurations in disk-backed modes
- Implement OAuth2 token endpoint at `/oauth2/v4/token` returning valid-looking mock tokens
- Keep memory mode unchanged (its contract is ephemeral storage)

**Non-Goals:**
- Real OAuth2 validation — the endpoint returns mock tokens for any input
- Changing the WAL format — we extend the existing JSON-line-based format
- gRPC API support
- Real Pub/Sub integration for notifications (storage-only, per existing design)

## Decisions

### Binary content: Save alongside JSON in a separate blobs directory

**Decision**: Store binary content as individual files under `<storagePath>/blobs/<bucket>/<object>#<generation>`, while keeping metadata JSON unchanged.

**Rationale**: The current `data.json` approach embeds `buckets` and `objects` as JSON maps. Marshaling multi-megabyte byte slices into JSON is inefficient and makes the file unreadable. A sidecar blob directory keeps metadata readable and content efficient.

**Alternatives considered**:
- Base64-encode content in JSON: Simple but bloated for large files, poor performance
- Single binary archive: Complex, harder to debug
- Separate file per blob: Clean, debuggable, matches real object storage patterns

### IAM and Notifications: Extend data.json

**Decision**: Add `iamPolicies` and `notifications` fields to the `persistentData` struct serialized to `data.json`.

**Rationale**: These are small, JSON-native structures (policy bindings, notification configs) that naturally serialize alongside bucket/object metadata. No separate file needed.

### Auth: Minimal OAuth2 endpoint

**Decision**: Implement a single `POST /oauth2/v4/token` handler in `internal/auth/handler.go` that accepts any `grant_type` and `client_id`/`client_secret` values and returns a fixed-format JSON response with valid-looking token fields.

**Rationale**: SDKs like the Python `google-cloud-storage` client attempt an OAuth2 token exchange during initialization. If the endpoint returns a valid response, the SDK proceeds. Real validation is explicitly out of scope (see `auth-emulation` spec Non-Goals). The endpoint is purely for SDK compatibility, not security.

### WAL: Extend operation types for IAM and notifications

**Decision**: Add `set_iam_policy`, `create_notification`, `delete_notification` operation types to the WAL format. Replay these during recovery and include them in compaction.

**Rationale**: The WAL already has typed operations for buckets and objects. Extending the pattern is natural. Compaction already saves a `snapshot.json` — we extend that to include IAM and notification state.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Blob files accumulate and aren't cleaned up on object delete | Delete blob file alongside object deletion in persistent/hybrid/WAL |
| Large files slow down WAL compaction | Compaction already snapshot-based; binary content goes to blob dir, not WAL lines |
| Auth endpoint could confuse users into thinking it validates credentials | Document clearly in code that it's a passthrough, match the auth-emulation spec language |
| Existing data.json files without iamPolicies/notifications fields | Gracefully handle missing fields on load (treat as empty maps) |

## Open Questions

None.
