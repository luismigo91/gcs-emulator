## ADDED Requirements

### Requirement: Memory storage mode
The emulator SHALL support a memory-only storage mode where all data is kept in RAM and lost when the process terminates.

#### Scenario: Default to memory mode
- **WHEN** the emulator starts without `GCP_EMULATOR_STORAGE_MODE` configured
- **THEN** the storage backend initializes in memory mode

#### Scenario: Store and retrieve in memory
- **WHEN** a bucket is created and an object is uploaded while in memory mode
- **THEN** the bucket and object are immediately accessible via API calls

#### Scenario: Data loss on restart
- **WHEN** the emulator is stopped and restarted in memory mode
- **THEN** all previously stored buckets and objects are absent

### Requirement: Persistent storage mode
The emulator SHALL support a persistent storage mode where all data is saved to disk on graceful shutdown and loaded on startup.

#### Scenario: Save on graceful shutdown
- **WHEN** the emulator receives a SIGTERM or SIGINT signal while in persistent mode
- **THEN** all buckets and objects are serialized to disk at `GCP_EMULATOR_STORAGE_PATH` before exit

#### Scenario: Load on startup
- **WHEN** the emulator starts in persistent mode and a data directory exists at `GCP_EMULATOR_STORAGE_PATH`
- **THEN** all previously saved buckets and objects are loaded and accessible

#### Scenario: No data on first start
- **WHEN** the emulator starts in persistent mode with no existing data directory
- **THEN** the emulator initializes with empty storage and creates the data directory

### Requirement: Hybrid storage mode
The emulator SHALL support a hybrid storage mode where data is kept in RAM for performance and asynchronously flushed to disk at regular intervals.

#### Scenario: Async flush interval
- **WHEN** the emulator runs in hybrid mode
- **THEN** storage state is flushed to disk every 5 seconds in a background goroutine

#### Scenario: Data survives restart in hybrid mode
- **WHEN** the emulator is restarted in hybrid mode after running with data
- **THEN** data from the last flush cycle is loaded and accessible

#### Scenario: Flush on graceful shutdown
- **WHEN** the emulator receives a shutdown signal while in hybrid mode
- **THEN** a final flush is performed before exit to minimize data loss

### Requirement: WAL (Write-Ahead Log) storage mode
The emulator SHALL support a WAL storage mode where every mutation is logged to disk before being applied to in-memory state.

#### Scenario: Write to WAL before applying
- **WHEN** a bucket or object mutation occurs in WAL mode
- **THEN** the operation is appended to the WAL file on disk before the in-memory state is updated

#### Scenario: Recovery from WAL
- **WHEN** the emulator starts in WAL mode with an existing WAL file
- **THEN** the WAL is replayed to reconstruct the in-memory state

#### Scenario: WAL compaction
- **WHEN** the WAL file exceeds 10MB or the emulator performs a graceful shutdown
- **THEN** the WAL is compacted into a snapshot file and the WAL is truncated

### Requirement: Per-service storage override
The emulator SHALL allow configuring storage mode per service, overriding the global setting.

#### Scenario: Override via environment variable
- **WHEN** `GCP_EMULATOR_SERVICES_GCS_STORAGE_MODE` is set to `persistent` while global `GCP_EMULATOR_STORAGE_MODE` is `memory`
- **THEN** the GCS service uses persistent mode while other services use memory mode

### Requirement: Storage backend interface
The emulator SHALL define a Go interface that all storage backends implement, enabling pluggable backend implementations.

#### Scenario: Backend interface contract
- **WHEN** a new storage backend is implemented
- **THEN** it must satisfy the `Backend` interface with methods: `CreateBucket`, `GetBucket`, `ListBuckets`, `DeleteBucket`, `CreateObject`, `GetObject`, `ListObjects`, `DeleteObject`, `UpdateObject`

#### Scenario: Swap backend at startup
- **WHEN** the storage mode configuration changes between runs
- **THEN** the emulator initializes the corresponding backend implementation without code changes

### Requirement: Configurable storage path
The emulator SHALL allow configuring the disk storage path via environment variable.

#### Scenario: Custom storage path
- **WHEN** `GCP_EMULATOR_STORAGE_PATH` is set to `/tmp/gcp-emulator-data`
- **THEN** persistent, hybrid, and WAL modes use that directory for disk operations

#### Scenario: Default storage path
- **WHEN** `GCP_EMULATOR_STORAGE_PATH` is not set
- **THEN** the emulator uses `./data` relative to the working directory
