## ADDED Requirements

### Requirement: Bucket CRUD operations
The emulator SHALL support creating, reading, updating, listing, and deleting GCS buckets via both JSON API and XML API endpoints.

#### Scenario: Create bucket via JSON API
- **WHEN** a POST request is sent to `/storage/v1/b` with a JSON body containing `{"name": "my-bucket"}`
- **THEN** the emulator creates the bucket and returns a JSON response with `kind: "storage#bucket"`, the bucket name, and HTTP 200

#### Scenario: List buckets via JSON API
- **WHEN** a GET request is sent to `/storage/v1/b`
- **THEN** the emulator returns a JSON response with `kind: "storage#buckets"` and an `items` array containing all buckets in the current project

#### Scenario: Get bucket metadata via JSON API
- **WHEN** a GET request is sent to `/storage/v1/b/{bucket}`
- **THEN** the emulator returns the bucket metadata as JSON with HTTP 200, or a GCP-formatted error with HTTP 404 if the bucket does not exist

#### Scenario: Delete bucket via JSON API
- **WHEN** a DELETE request is sent to `/storage/v1/b/{bucket}` for an empty bucket
- **THEN** the emulator deletes the bucket and returns HTTP 204

#### Scenario: Delete non-empty bucket fails
- **WHEN** a DELETE request is sent to `/storage/v1/b/{bucket}` for a bucket containing objects
- **THEN** the emulator returns a GCP-formatted error with HTTP 409

### Requirement: Object CRUD operations
The emulator SHALL support creating, reading, listing, updating, and deleting GCS objects via both JSON API and XML API endpoints.

#### Scenario: Upload object via JSON API
- **WHEN** a POST request with multipart/form-data is sent to `/upload/storage/v1/b/{bucket}/o` with object content
- **THEN** the emulator stores the object and returns JSON metadata with `kind: "storage#object"`, name, bucket, size, and HTTP 200

#### Scenario: Download object via JSON API
- **WHEN** a GET request is sent to `/storage/v1/b/{bucket}/o/{object}?alt=media`
- **THEN** the emulator returns the object content with the correct `Content-Type` and HTTP 200, or HTTP 404 if not found

#### Scenario: List objects via JSON API
- **WHEN** a GET request is sent to `/storage/v1/b/{bucket}/o`
- **THEN** the emulator returns a JSON response with `kind: "storage#objects"` and an `items` array containing object metadata

#### Scenario: List objects with prefix and delimiter
- **WHEN** a GET request is sent to `/storage/v1/b/{bucket}/o?prefix=logs/&delimiter=/`
- **THEN** the emulator returns objects matching the prefix and `prefixes` array for common prefix directories

#### Scenario: Delete object via JSON API
- **WHEN** a DELETE request is sent to `/storage/v1/b/{bucket}/o/{object}`
- **THEN** the emulator deletes the object and returns HTTP 204, or HTTP 404 if not found

#### Scenario: Upload object via XML API
- **WHEN** a PUT request with body content is sent to `/{bucket}/{object}`
- **THEN** the emulator stores the object and returns HTTP 200

#### Scenario: Download object via XML API
- **WHEN** a GET request is sent to `/{bucket}/{object}`
- **THEN** the emulator returns the object content with HTTP 200

### Requirement: Resumable uploads
The emulator SHALL support GCS resumable upload protocol for large objects, allowing clients to upload in chunks and resume interrupted uploads.

#### Scenario: Initiate resumable upload
- **WHEN** a POST request with `X-Goog-Upload-Command: start` and `X-Goog-Upload-Header-Content-Length` is sent to `/resumable/upload/storage/v1/b/{bucket}/o`
- **THEN** the emulator returns HTTP 200 with an `X-Goog-Upload-URL` header containing a session ID

#### Scenario: Upload chunk
- **WHEN** a PUT request with `X-Goog-Upload-Command: upload` and `Content-Range` header is sent to the resumable upload URL
- **THEN** the emulator stores the chunk and returns HTTP 308 with `X-Goog-Upload-Chunk-Granularity`

#### Scenario: Finalize resumable upload
- **WHEN** a PUT request with `X-Goog-Upload-Command: upload, finalize` and the final chunk is sent
- **THEN** the emulator assembles the object and returns JSON metadata with HTTP 200

#### Scenario: Query upload status
- **WHEN** a PUT request with `X-Goog-Upload-Command: query` is sent to the resumable upload URL
- **THEN** the emulator returns HTTP 308 with the current byte range received

### Requirement: Object compose
The emulator SHALL support composing multiple objects into a single composite object.

#### Scenario: Compose objects
- **WHEN** a POST request is sent to `/storage/v1/b/{bucket}/o/{destination}/compose` with a JSON body listing source objects
- **THEN** the emulator creates a new object containing the concatenated content of source objects and returns its metadata

### Requirement: Object copy and rewrite
The emulator SHALL support copying and rewriting objects within and across buckets.

#### Scenario: Copy object
- **WHEN** a POST request is sent to `/storage/v1/b/{sourceBucket}/o/{sourceObject}/copyTo/b/{destBucket}/o/{destObject}`
- **THEN** the emulator creates a copy of the object in the destination bucket and returns its metadata

#### Scenario: Rewrite object
- **WHEN** a POST request is sent to `/storage/v1/b/{sourceBucket}/o/{sourceObject}/rewriteTo/b/{destBucket}/o/{destObject}`
- **THEN** the emulator returns a rewrite response with `done: true` and the rewritten object metadata

### Requirement: Object versioning
The emulator SHALL support GCS object versioning, allowing multiple generations of an object to coexist in a versioned bucket.

#### Scenario: Enable versioning on bucket
- **WHEN** a PATCH request updates a bucket's `versioning.enabled` field to `true`
- **THEN** the bucket is marked as versioned and subsequent overwrites create new generations

#### Scenario: Upload new generation of existing object
- **WHEN** an object is uploaded to a versioned bucket with the same name as an existing object
- **THEN** the emulator creates a new generation with a unique `generation` number while preserving the previous generation

#### Scenario: Get specific generation
- **WHEN** a GET request is sent to `/storage/v1/b/{bucket}/o/{object}?generation={generation}`
- **THEN** the emulator returns the specific generation of the object

#### Scenario: Delete specific generation
- **WHEN** a DELETE request is sent to `/storage/v1/b/{bucket}/o/{object}?generation={generation}`
- **THEN** the emulator deletes only the specified generation, leaving other generations intact

### Requirement: Object metadata and custom metadata
The emulator SHALL support GCS object metadata including content-type, content-encoding, cache-control, and custom metadata key-value pairs.

#### Scenario: Set content-type on upload
- **WHEN** an object is uploaded with `Content-Type: application/json`
- **THEN** the stored object metadata includes `contentType: "application/json"`

#### Scenario: Set custom metadata
- **WHEN** an object is uploaded with `x-goog-meta-key: value` header
- **THEN** the stored object metadata includes `metadata: {"key": "value"}`

#### Scenario: Update object metadata
- **WHEN** a PATCH request is sent to `/storage/v1/b/{bucket}/o/{object}` with updated metadata fields
- **THEN** the object metadata is updated and the response reflects the changes

### Requirement: Range requests for downloads
The emulator SHALL support HTTP Range headers for partial object downloads.

#### Scenario: Download byte range
- **WHEN** a GET request includes `Range: bytes=0-99` header
- **THEN** the emulator returns HTTP 206 with `Content-Range: bytes 0-99/{total}` and the first 100 bytes

#### Scenario: Invalid range
- **WHEN** a GET request includes `Range: bytes=1000-2000` for an object smaller than 1000 bytes
- **THEN** the emulator returns HTTP 416 Range Not Satisfiable

### Requirement: GCP error response format
The emulator SHALL return errors in the GCP JSON error format for all failed requests.

#### Scenario: Bucket not found error
- **WHEN** a request references a non-existent bucket
- **THEN** the response body contains `{"error": {"code": 404, "message": "Not Found", "errors": [...]}}` with HTTP 404

#### Scenario: Invalid request error
- **WHEN** a request has invalid parameters
- **THEN** the response body contains `{"error": {"code": 400, "message": "...", "errors": [...]}}` with HTTP 400

### Requirement: Bucket lifecycle rules
The emulator SHALL support defining and evaluating bucket lifecycle rules for object management.

#### Scenario: Create lifecycle rule
- **WHEN** a PATCH request sets `lifecycle.rule` on a bucket with conditions and actions
- **THEN** the bucket stores the lifecycle rules and returns the updated configuration

#### Scenario: Lifecycle rule evaluation
- **WHEN** an object meets the conditions of a lifecycle rule (e.g., age > N days)
- **THEN** the emulator applies the configured action (delete, set storage class) when listing or accessing the object

### Requirement: Bucket IAM policies
The emulator SHALL support getting and setting IAM policies on buckets.

#### Scenario: Get IAM policy
- **WHEN** a GET request is sent to `/storage/v1/b/{bucket}/iam`
- **THEN** the emulator returns the bucket's IAM policy with bindings

#### Scenario: Set IAM policy
- **WHEN** a PUT request is sent to `/storage/v1/b/{bucket}/iam` with a policy body
- **THEN** the emulator updates the bucket's IAM policy and returns the new policy

### Requirement: Bucket notifications
The emulator SHALL support creating, listing, and deleting GCS notification configurations on buckets.

#### Scenario: Create notification
- **WHEN** a POST request is sent to `/storage/v1/b/{bucket}/notificationConfigs` with topic and payload format
- **THEN** the emulator creates the notification configuration and returns it with a generated ID

#### Scenario: List notifications
- **WHEN** a GET request is sent to `/storage/v1/b/{bucket}/notificationConfigs`
- **THEN** the emulator returns all notification configurations for the bucket
