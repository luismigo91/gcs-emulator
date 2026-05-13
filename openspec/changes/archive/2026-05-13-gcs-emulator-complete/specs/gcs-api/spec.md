## MODIFIED Requirements

### Requirement: Response Headers

The emulator SHALL return standard GCS response headers for object operations with actual values.

#### Scenario: Object download headers

**WHEN** a client downloads an object
**THEN** the response SHALL include:
- `X-Goog-Generation`: the object's generation
- `X-Goog-Hash`: `crc32c=<crc32c>,md5=<md5>`
- `X-Goog-Stored-Content-Length`: the stored content size
- `ETag`: a quoted hash value (base64-encoded generation)
- `Last-Modified`: the object's updated time in HTTP date format
- `Accept-Ranges: bytes`

### Requirement: Complete Object Metadata

The emulator SHALL support all standard GCS object metadata fields including correct ProjectNumber.

#### Scenario: Object includes all metadata fields

**WHEN** a client creates or retrieves an object
**THEN** the object SHALL include fields:
- `contentEncoding`
- `contentDisposition`
- `contentLanguage`
- `cacheControl`
- `customTime`
- `etag`
- `storageClass`
- `timeStorageClassUpdated`

#### Scenario: Bucket includes project number

**WHEN** a client creates or retrieves a bucket
**THEN** the bucket SHALL include `projectNumber` as a numeric string derived from the request context
