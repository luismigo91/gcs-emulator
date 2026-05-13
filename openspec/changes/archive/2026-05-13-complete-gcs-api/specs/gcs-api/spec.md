## ADDED Requirements

### Requirement: XML API DELETE Bucket
The emulator SHALL support deleting a bucket via the XML API with `DELETE /{bucket}`.

#### Scenario: Delete empty bucket via XML
- **WHEN** `DELETE /my-bucket` is called via XML API
- **AND** the bucket exists and is empty
- **THEN** the bucket is deleted and HTTP 204 No Content is returned

#### Scenario: Delete non-empty bucket via XML
- **WHEN** `DELETE /my-bucket` is called via XML API
- **AND** the bucket contains objects
- **THEN** HTTP 409 Conflict is returned

#### Scenario: Delete non-existent bucket via XML
- **WHEN** `DELETE /nonexistent` is called via XML API
- **THEN** HTTP 404 Not Found is returned

## MODIFIED Requirements

### Requirement: Response Headers
The emulator SHALL return standard GCS response headers for object operations.

#### Scenario: Object download headers

- **WHEN** a client downloads an object
- **THEN** the response SHALL include:
- `X-Goog-Generation`: the object's generation
- `X-Goog-Hash`: `crc32c=<crc32c>,md5=<md5>`
- `X-Goog-Stored-Content-Length`: the stored content size
- `ETag`: the object's ETag
- `Last-Modified`: the object's updated time in HTTP date format
- `Accept-Ranges: bytes`

#### Scenario: Object metadata headers

- **WHEN** a client requests object metadata
- **THEN** the response SHALL include:
- `Accept-Ranges: bytes`

#### Scenario: Custom metadata headers

- **GIVEN** an object has custom metadata `{"x-custom-key": "value"}`
- **WHEN** a client downloads the object
- **THEN** the response SHALL include `X-Goog-Meta-X-Custom-Key: value`

#### Scenario: Health check response
- **WHEN** `GET /-/health` is called
- **THEN** the response SHALL be HTTP 200 with JSON body `{"status":"healthy","services":{"gcs":"available"}}`
