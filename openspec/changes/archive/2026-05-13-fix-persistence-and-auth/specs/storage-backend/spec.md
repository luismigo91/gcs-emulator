## MODIFIED Requirements

### Requirement: Persistent storage mode
The emulator SHALL support a persistent storage mode where all data — including object binary content, IAM policies, and notification configurations — is saved to disk on every mutation and loaded on startup.

#### Scenario: Save on graceful shutdown
- **WHEN** the emulator receives a SIGTERM or SIGINT signal while in persistent mode
- **THEN** all buckets, objects with their binary content, IAM policies, and notifications are serialized to disk at `GCP_EMULATOR_STORAGE_PATH` before exit

#### Scenario: Load on startup
- **WHEN** the emulator starts in persistent mode and a data directory exists at `GCP_EMULATOR_STORAGE_PATH`
- **THEN** all previously saved buckets, objects with their binary content, IAM policies, and notification configurations are loaded and fully accessible

#### Scenario: Object content survives restart
- **WHEN** an object with binary content "hello world" is created in persistent mode and the emulator is restarted
- **THEN** downloading the object returns the exact original content "hello world"

#### Scenario: IAM policies survive restart
- **WHEN** an IAM policy is set on a bucket in persistent mode and the emulator is restarted
- **THEN** retrieving the bucket's IAM policy returns the same policy

#### Scenario: Notifications survive restart
- **WHEN** a notification configuration is created in persistent mode and the emulator is restarted
- **THEN** listing notifications for the bucket returns the same configuration

#### Scenario: No data on first start
- **WHEN** the emulator starts in persistent mode with no existing data directory
- **THEN** the emulator initializes with empty storage and creates the data directory

### Requirement: Hybrid storage mode
The emulator SHALL support a hybrid storage mode where data is kept in RAM for performance and asynchronously flushed to disk at regular intervals, including object binary content, IAM policies, and notification configurations.

#### Scenario: Async flush interval
- **WHEN** the emulator runs in hybrid mode
- **THEN** storage state including binary content, IAM policies, and notifications is flushed to disk every 5 seconds in a background goroutine

#### Scenario: Data survives restart in hybrid mode
- **WHEN** the emulator is restarted in hybrid mode after running with data
- **THEN** data from the last flush cycle — including object content, IAM policies, and notifications — is loaded and accessible

#### Scenario: Flush on graceful shutdown
- **WHEN** the emulator receives a shutdown signal while in hybrid mode
- **THEN** a final flush is performed before exit to minimize data loss

### Requirement: WAL (Write-Ahead Log) storage mode
The emulator SHALL support a WAL storage mode where every mutation — including IAM policy changes and notification CRUD — is logged to disk before being applied to in-memory state.

#### Scenario: Write to WAL before applying
- **WHEN** an IAM policy or notification mutation occurs in WAL mode
- **THEN** the operation is appended to the WAL file on disk before the in-memory state is updated

#### Scenario: Recovery from WAL
- **WHEN** the emulator starts in WAL mode with an existing WAL file containing IAM and notification operations
- **THEN** the WAL is replayed to reconstruct the full in-memory state including IAM policies and notifications

#### Scenario: WAL compaction
- **WHEN** the WAL file exceeds 10MB or the emulator performs a graceful shutdown
- **THEN** the WAL is compacted into a snapshot file that includes IAM policies and notification configurations, and the WAL is truncated

### Requirement: Object content persistence
The emulator SHALL persist object binary content to disk separately from metadata JSON to avoid encoding bloat and keep the metadata file human-readable.

#### Scenario: Binary content stored as blobs
- **WHEN** an object is created or updated in a disk-backed mode
- **THEN** its binary content SHALL be written to `<storagePath>/blobs/<bucket>/<object>#<generation>` as raw bytes

#### Scenario: Blob cleaned up on object delete
- **WHEN** an object is deleted in a disk-backed mode
- **THEN** its associated blob file SHALL be removed from disk

#### Scenario: Backward-compatible loading
- **WHEN** an existing `data.json` file without `iamPolicies` or `notifications` fields is loaded
- **THEN** the emulator SHALL initialize those fields as empty maps and continue normally
