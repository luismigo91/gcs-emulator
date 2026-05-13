## 1. Model & Backend Foundation

- [x] 1.1 Change `Object.Generation` from `string` to `int64` in `internal/model/object.go`
- [x] 1.2 Add missing metadata fields to Object struct: `ContentEncoding`, `ContentDisposition`, `ContentLanguage`, `CacheControl`, `CustomTime`, `ETag`, `TimeStorageClassUpdated`
- [x] 1.3 Add `Conditions` struct to `internal/backend/backend.go` for conditional request validation
- [x] 1.4 Update `Backend` interface methods to accept `Conditions` parameter where applicable
- [x] 1.5 Add `ListObjectsResponse` struct with `Items`, `Prefixes`, `NextPageToken` fields

## 2. Checksum Utilities

- [x] 2.1 Implement CRC32C calculation function (Google polynomial) in `internal/util/checksum.go`
- [x] 2.2 Implement MD5 calculation function
- [x] 2.3 Implement base64 encoding helpers for checksums

## 3. Backend — Generations & Versioning

- [x] 3.1 Update `MemoryBackend` to use int64 generations from `time.Now().UnixNano()` with counter fallback for collisions
- [x] 3.2 Implement versioning: when versioning is enabled, old objects are preserved on overwrite
- [x] 3.3 Update `GetObject` to support generation parameter for retrieving specific versions
- [x] 3.4 Update `ListObjects` to support `Versions` flag (show all generations vs only live)
- [x] 3.5 Update `DeleteObject` to handle versioning (soft delete when versioning enabled without generation param)
- [x] 3.6 Propagate generation changes to `PersistentBackend`, `HybridBackend`, `WALBackend`

## 4. Backend — Conditional Requests

- [x] 4.1 Implement condition validation in `MemoryBackend`: `DoesNotExist`, `GenerationMatch`, `GenerationNotMatch`, `MetagenerationMatch`, `MetagenerationNotMatch`
- [x] 4.2 Return appropriate errors for failed conditions (map to HTTP 412)

## 5. Backend — Resumable Upload Sessions

- [x] 5.1 Create `UploadSession` struct in `internal/backend/upload.go` with session ID, bucket, object name, chunks, offset, expiry
- [x] 5.2 Implement session store with `sync.Map` and 7-day expiry
- [x] 5.3 Implement chunk accumulation: store chunks by offset, merge on finalize
- [x] 5.4 Implement session cleanup goroutine for expired sessions

## 6. Backend — Compose Objects

- [x] 6.1 Implement `ComposeObjects` in `MemoryBackend`: concatenate source object contents, calculate checksums, set componentCount
- [x] 6.2 Validate max 32 source objects
- [x] 6.3 Propagate to other backend implementations

## 7. API — List Objects with CommonPrefixes

- [x] 7.1 Implement prefix computation logic: given prefix + delimiter, extract common prefixes from object names
- [x] 7.2 Support `startOffset`, `endOffset`, `includeTrailingDelimiter` parameters
- [x] 7.3 Support `maxResults` and `pageToken` pagination
- [x] 7.4 Return `prefixes` array in JSON response alongside `items`

## 8. API — Range Requests

- [x] 8.1 Implement `parseRange` function for parsing `Range: bytes=X-Y` header
- [x] 8.2 Implement range handling: return HTTP 206 with `Content-Range` header
- [x] 8.3 Handle suffix ranges (`bytes=-N`), open-ended ranges (`bytes=N-`)
- [x] 8.4 Handle unsatisfiable ranges (HTTP 416)
- [x] 8.5 Handle invalid ranges (return full content per GCS behavior)

## 9. API — Response Headers

- [x] 9.1 Create `setObjectResponseHeaders` helper function
- [x] 9.2 Add headers: `X-Goog-Generation`, `X-Goog-Hash`, `X-Goog-Stored-Content-Length`, `ETag`, `Last-Modified`, `Accept-Ranges`
- [x] 9.3 Add custom metadata headers: `X-Goog-Meta-*`
- [x] 9.4 Apply headers to download, metadata, and upload responses

## 10. API — Multipart Upload

- [x] 10.1 Detect `Content-Type: multipart/related` in upload handlers
- [x] 10.2 Parse multipart body: extract JSON metadata from first part, content from second part
- [x] 10.3 Create object with parsed metadata and content

## 11. API — Resumable Upload Handlers

- [x] 11.1 Implement session creation: generate session ID, store in session store, return upload URL
- [x] 11.2 Implement chunk upload: validate session, store chunk at offset, update offset
- [x] 11.3 Implement finalize: merge chunks, create object, clean up session
- [x] 11.4 Implement status query: return current offset and status

## 12. API — Compose Objects Handler

- [x] 12.1 Implement `POST /storage/v1/b/{bucket}/o/{object}/compose` handler
- [x] 12.2 Parse compose request body (sourceObjects array, optional destination metadata)
- [x] 12.3 Call backend ComposeObjects, return response

## 13. API — Conditional Request Middleware

- [x] 13.1 Parse query parameters: `ifGenerationMatch`, `ifGenerationNotMatch`, `ifMetagenerationMatch`, `ifMetagenerationNotMatch`
- [x] 13.2 Pass conditions to backend methods
- [x] 13.3 Map backend condition errors to HTTP 412 responses

## 14. API — Object Metadata Updates

- [x] 14.1 Update `PATCH /storage/v1/b/{bucket}/o/{object}` handler to support all metadata fields
- [x] 14.2 Update `ObjectUpdateAttrs` struct with new fields

## 15. Testing

- [x] 15.1 Write unit tests for checksum calculation (CRC32C, MD5)
- [x] 15.2 Write unit tests for range parsing
- [x] 15.3 Write unit tests for prefix computation
- [x] 15.4 Write unit tests for conditional request validation
- [x] 15.5 Write integration tests for resumable upload flow
- [x] 15.6 Write integration tests for compose objects
- [x] 15.7 Write integration tests for versioning/generations
- [x] 15.8 Write integration tests for list objects with prefixes and pagination
