# gcs-api

## Requirements

### Requirement: List Objects with Common Prefixes

The emulator SHALL return `prefixes` array in list objects responses when a delimiter is specified, enabling SDKs to navigate hierarchical object structures.

#### Scenario: List objects with delimiter

**Given** a bucket contains objects: `logs/2024/jan.txt`, `logs/2024/feb.txt`, `logs/2025/mar.txt`
**When** a client sends `GET /storage/v1/b/{bucket}/o?prefix=logs/&delimiter=/`
**Then** the response SHALL include:
```json
{
  "kind": "storage#objects",
  "items": [],
  "prefixes": ["logs/2024/", "logs/2025/"]
}
```

#### Scenario: List objects with includeTrailingDelimiter

**Given** a bucket contains objects: `logs/2024/jan.txt`
**When** a client sends `GET /storage/v1/b/{bucket}/o?prefix=logs/&delimiter=/&includeTrailingDelimiter=true`
**Then** the response SHALL include the prefix object `logs/2024/` in items AND `logs/2024/` in prefixes

#### Scenario: List objects with pagination

**Given** a bucket contains 100 objects
**When** a client sends `GET /storage/v1/b/{bucket}/o?maxResults=10`
**Then** the response SHALL include at most 10 items and a `nextPageToken` for the next page

#### Scenario: List objects with startOffset and endOffset

**Given** a bucket contains objects: `a.txt`, `b.txt`, `c.txt`, `d.txt`
**When** a client sends `GET /storage/v1/b/{bucket}/o?startOffset=b&endOffset=d`
**Then** the response SHALL include only `b.txt` and `c.txt`

### Requirement: Range Requests

The emulator SHALL support HTTP Range headers for partial object downloads, returning HTTP 206 Partial Content with proper Content-Range headers.

#### Scenario: Partial download with byte range

**Given** an object with 1000 bytes of content
**When** a client sends `GET /storage/v1/b/{bucket}/o/{object}?alt=media` with `Range: bytes=0-99`
**Then** the response SHALL be HTTP 206 with:
- `Content-Range: bytes 0-99/1000`
- `Content-Length: 100`
- The first 100 bytes of content

#### Scenario: Range from offset to end

**Given** an object with 1000 bytes of content
**When** a client sends `GET /storage/v1/b/{bucket}/o/{object}?alt=media` with `Range: bytes=500-`
**Then** the response SHALL be HTTP 206 with bytes 500-999

#### Scenario: Suffix range (last N bytes)

**Given** an object with 1000 bytes of content
**When** a client sends `GET /storage/v1/b/{bucket}/o/{object}?alt=media` with `Range: bytes=-100`
**Then** the response SHALL be HTTP 206 with the last 100 bytes

#### Scenario: Range beyond content length

**Given** an object with 100 bytes of content
**When** a client sends `GET /storage/v1/b/{bucket}/o/{object}?alt=media` with `Range: bytes=200-300`
**Then** the response SHALL be HTTP 416 with an XML error response

#### Scenario: Invalid range

**Given** an object with content
**When** a client sends a request with an invalid Range header format
**Then** the emulator SHALL return the full content (GCS behavior: ignore invalid ranges)

### Requirement: Object Generations and Versioning

The emulator SHALL support object generations as int64 values and bucket versioning for preserving object history.

#### Scenario: Create object generates unique generation

**When** a client creates an object in a bucket
**Then** the object SHALL have a `generation` field set to a unique int64 value (nanosecond timestamp)

#### Scenario: Overwrite creates new generation

**Given** an object exists with generation 100
**When** a client creates an object with the same name
**Then** a new object SHALL be created with a different generation (e.g., 200)
**And** the old generation SHALL still be accessible if versioning is enabled

#### Scenario: Delete without generation (versioning enabled)

**Given** versioning is enabled on a bucket
**And** an object exists with generation 100
**When** a client deletes the object without specifying generation
**Then** the object's live version SHALL be removed
**And** the generation 100 version SHALL still be accessible via `?generation=100`

#### Scenario: Delete with generation

**Given** an object exists with generations 100 and 200
**When** a client deletes the object with `?generation=100`
**Then** only generation 100 SHALL be removed
**And** generation 200 SHALL still exist

### Requirement: Conditional Requests

The emulator SHALL support conditional request headers for optimistic concurrency control.

#### Scenario: ifGenerationMatch succeeds

**Given** an object exists with generation 100
**When** a client sends a request with `?ifGenerationMatch=100`
**Then** the operation SHALL proceed

#### Scenario: ifGenerationMatch fails

**Given** an object exists with generation 100
**When** a client sends a request with `?ifGenerationMatch=200`
**Then** the response SHALL be HTTP 412 Precondition Failed

#### Scenario: ifGenerationNotMatch succeeds

**Given** an object exists with generation 100
**When** a client sends a request with `?ifGenerationNotMatch=200`
**Then** the operation SHALL proceed

#### Scenario: ifMetagenerationMatch

**Given** a bucket exists with metageneration 5
**When** a client sends a request with `?ifMetagenerationMatch=5`
**Then** the operation SHALL proceed
**When** a client sends a request with `?ifMetagenerationMatch=3`
**Then** the response SHALL be HTTP 412

#### Scenario: ifGenerationMatch on create (DoesNotExist)

**Given** an object does not exist
**When** a client creates with `?ifGenerationMatch=0`
**Then** the operation SHALL proceed
**When** the object exists and a client creates with `?ifGenerationMatch=0`
**Then** the response SHALL be HTTP 412

### Requirement: Compose Objects

The emulator SHALL support composing up to 32 source objects into a single destination object.

#### Scenario: Compose two objects

**Given** objects `part1.txt` (content "hello") and `part2.txt` (content " world") exist in a bucket
**When** a client sends `POST /storage/v1/b/{bucket}/o/result.txt/compose` with:
```json
{"sourceObjects": [{"name": "part1.txt"}, {"name": "part2.txt"}]}
```
**Then** a new object `result.txt` SHALL be created with content "hello world"
**And** `componentCount` SHALL be 2

#### Scenario: Compose exceeds maximum sources

**When** a client sends a compose request with 33 source objects
**Then** the response SHALL be HTTP 400 with an error message about exceeding the 32-object limit

#### Scenario: Compose with destination metadata

**Given** source objects exist
**When** a client sends a compose request with a `destination` object containing `contentType` and `metadata`
**Then** the composed object SHALL have the specified contentType and metadata

### Requirement: Response Headers

The emulator SHALL return standard GCS response headers for object operations.

#### Scenario: Object download headers

**When** a client downloads an object
**Then** the response SHALL include:
- `X-Goog-Generation`: the object's generation
- `X-Goog-Hash`: `crc32c=<crc32c>,md5=<md5>`
- `X-Goog-Stored-Content-Length`: the stored content size
- `ETag`: the object's ETag
- `Last-Modified`: the object's updated time in HTTP date format
- `Accept-Ranges: bytes`

#### Scenario: Object metadata headers

**When** a client requests object metadata
**Then** the response SHALL include:
- `Accept-Ranges: bytes`

#### Scenario: Custom metadata headers

**Given** an object has custom metadata `{"x-custom-key": "value"}`
**When** a client downloads the object
**Then** the response SHALL include `X-Goog-Meta-X-Custom-Key: value`

### Requirement: Multipart Upload

The emulator SHALL support multipart/form-data uploads that include both object metadata and content in a single request.

#### Scenario: Multipart upload with metadata

**When** a client sends `POST /upload/storage/v1/b/{bucket}/o?name=test.txt` with `Content-Type: multipart/related` containing:
- Part 1: JSON metadata `{"contentType": "text/plain", "metadata": {"key": "value"}}`
- Part 2: Object content "hello world"
**Then** an object SHALL be created with the specified metadata and content

### Requirement: Resumable Upload Sessions

The emulator SHALL support resumable upload sessions that persist across multiple requests and accumulate chunks.

#### Scenario: Create resumable upload session

**When** a client sends `POST /resumable/upload/storage/v1/b/{bucket}/o?name=test.txt` with `X-Goog-Upload-Command: start`
**Then** the response SHALL include:
- `X-Goog-Upload-URL`: a unique session URL
- `X-Goog-Upload-Status: active`

#### Scenario: Upload chunk

**Given** an active resumable upload session
**When** a client sends a PUT request to the session URL with `Content-Range: bytes 0-99/1000`
**Then** the response SHALL include `X-Goog-Upload-Status: active` and `X-Goog-Upload-Offset: 100`

#### Scenario: Finalize resumable upload

**Given** an active resumable upload session with accumulated chunks
**When** a client sends the final chunk with `X-Goog-Upload-Command: upload, finalize`
**Then** the object SHALL be created with the combined content
**And** the response SHALL include `X-Goog-Upload-Status: final`

#### Scenario: Query upload status

**Given** an active resumable upload session
**When** a client sends a request with `X-Goog-Upload-Command: query`
**Then** the response SHALL include `X-Goog-Upload-Status: active` and the current offset

### Requirement: Checksums

The emulator SHALL calculate and return CRC32C and MD5 checksums for uploaded objects.

#### Scenario: Checksums calculated on upload

**When** a client uploads an object with content "hello world"
**Then** the object SHALL have:
- `crc32c`: base64-encoded CRC32C checksum
- `md5Hash`: base64-encoded MD5 hash

#### Scenario: Checksums returned in download headers

**When** a client downloads an object
**Then** the `X-Goog-Hash` header SHALL include both `crc32c=<value>` and `md5=<value>`

### Requirement: Complete Object Metadata

The emulator SHALL support all standard GCS object metadata fields.

#### Scenario: Object includes all metadata fields

**When** a client creates or retrieves an object
**Then** the object SHALL include fields:
- `contentEncoding`
- `contentDisposition`
- `contentLanguage`
- `cacheControl`
- `customTime`
- `etag`
- `storageClass`
- `timeStorageClassUpdated`

#### Scenario: Update object metadata fields

**When** a client sends `PATCH /storage/v1/b/{bucket}/o/{object}` with `{"contentEncoding": "gzip", "cacheControl": "public, max-age=3600"}`
**Then** the object's metadata SHALL be updated accordingly
