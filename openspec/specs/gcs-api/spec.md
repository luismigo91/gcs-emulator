# gcs-api

## Requirements

### Requirement: Bucket Operations
The emulator SHALL support bucket CRUD operations via the JSON API.

#### Scenario: Create and get bucket
- **WHEN** POST /storage/v1/b with bucket name
- **THEN** bucket is created and returned via GET

### Requirement: Object Operations
The emulator SHALL support object CRUD, upload, download, compose, copy, and rewrite.

#### Scenario: Upload and download
- **WHEN** object is uploaded via simple/multipart/resumable upload
- **THEN** object is accessible via GET with alt=media

### Requirement: List Objects with Common Prefixes
The emulator SHALL return prefixes array when delimiter is specified.

#### Scenario: List with delimiter
- **GIVEN** objects logs/2024/jan.txt and logs/2025/mar.txt
- **WHEN** listing with prefix=logs/ and delimiter=/
- **THEN** prefixes [logs/2024/, logs/2025/] are returned

### Requirement: Range Requests
The emulator SHALL support HTTP Range headers for partial downloads.

#### Scenario: Partial download
- **WHEN** Range: bytes=0-99 is sent
- **THEN** HTTP 206 with Content-Range header and partial content

### Requirement: Object Generations and Versioning
The emulator SHALL support object generations and bucket versioning.

#### Scenario: Overwrite creates new generation
- **WHEN** object is overwritten with versioning enabled
- **THEN** old generation is preserved and accessible

### Requirement: Conditional Requests
The emulator SHALL support ifGenerationMatch, ifGenerationNotMatch, ifMetagenerationMatch, ifMetagenerationNotMatch.

#### Scenario: ifGenerationMatch succeeds
- **WHEN** request with ifGenerationMatch matching current generation
- **THEN** operation proceeds, otherwise HTTP 412

### Requirement: Compose Objects
The emulator SHALL compose up to 32 source objects into a destination object.

#### Scenario: Compose two objects
- **WHEN** POST /storage/v1/b/{bucket}/o/result.txt/compose with sources
- **THEN** result.txt has concatenated content and componentCount

### Requirement: Multipart Upload
The emulator SHALL support multipart/form-data uploads with metadata.

#### Scenario: Multipart with JSON metadata and content
- **WHEN** multipart/related upload with metadata JSON and content part
- **THEN** object is created with specified metadata

### Requirement: Resumable Upload Sessions
The emulator SHALL support resumable upload with start, chunk, finalize, query.

#### Scenario: Start and finalize resumable upload
- **WHEN** X-Goog-Upload-Command: start creates session
- **THEN** chunks can be uploaded and the session finalized into an object

### Requirement: Checksums
The emulator SHALL calculate CRC32C and MD5 checksums for uploaded objects.

#### Scenario: Checksums returned in download headers
- **WHEN** object is downloaded
- **THEN** X-Goog-Hash includes crc32c and md5 values

### Requirement: IAM Policies
The emulator SHALL support getIamPolicy, setIamPolicy, and testIamPermissions for buckets.

#### Scenario: Set and get IAM policy
- **WHEN** POST /storage/v1/b/{bucket}/iam with policy
- **THEN** policy is returned via GET

### Requirement: Bucket Notifications
The emulator SHALL support notification configuration CRUD.

#### Scenario: Create notification
- **WHEN** POST /storage/v1/b/{bucket}/notificationConfigs with topic and event types
- **THEN** configuration is stored and deliverable to Pub/Sub

### Requirement: Lifecycle Rules
The emulator SHALL support lifecycle rule configuration per bucket.

#### Scenario: Set lifecycle
- **WHEN** PATCH /storage/v1/b/{bucket}/lifecycle with rules
- **THEN** rules are stored, returned via GET, removable via DELETE

### Requirement: CORS Configuration
The emulator SHALL support CORS rules per bucket.

#### Scenario: Set CORS
- **WHEN** PATCH /storage/v1/b/{bucket}/cors with CORS rules
- **THEN** rules are stored and returned via GET

### Requirement: Bucket ACLs
The emulator SHALL support bucket-level ACL CRUD.

#### Scenario: Create and list ACLs
- **WHEN** POST /storage/v1/b/{bucket}/acl with entity and role
- **THEN** ACL is stored and visible in GET list

### Requirement: Object ACLs
The emulator SHALL support object-level ACL CRUD.

#### Scenario: Set object ACL
- **WHEN** POST /storage/v1/b/{bucket}/o/{object}/acl
- **THEN** ACL is stored and accessible

### Requirement: Default Object ACLs
The emulator SHALL support default object ACLs for buckets.

#### Scenario: Set default ACL
- **WHEN** POST /storage/v1/b/{bucket}/defaultObjectAcl
- **THEN** defaults are stored

### Requirement: Website Configuration
The emulator SHALL support bucket website config.

#### Scenario: Set website
- **WHEN** PUT /storage/v1/b/{bucket}/website
- **THEN** config is stored and returned

### Requirement: Encryption Configuration
The emulator SHALL support bucket encryption config.

#### Scenario: Set encryption
- **WHEN** PUT /storage/v1/b/{bucket}/encryption
- **THEN** config is stored

### Requirement: Signed URLs
The emulator SHALL generate and validate signed URLs.

#### Scenario: Accept signed URL
- **WHEN** GET with signed URL parameters
- **THEN** object is served without signature validation

### Requirement: XML API
The emulator SHALL support XML API for buckets and objects.

#### Scenario: XML upload and download
- **WHEN** PUT /{bucket}/{object} with content
- **THEN** object is stored and accessible via GET in XML format

### Requirement: Health Check
The emulator SHALL expose a health check endpoint.

#### Scenario: Health check
- **WHEN** GET /-/health
- **THEN** HTTP 200 with service status for all running services
