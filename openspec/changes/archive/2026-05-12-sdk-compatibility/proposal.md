## Why

Our GCS emulator currently covers ~35% of the real GCS API surface. fake-gcs-server (the established Go alternative) covers ~85%. The gap prevents our emulator from being a drop-in replacement for official Google Cloud SDKs (Go, Python, Node.js, Java). Critical failures include:

- **Python SDK**: Resumable uploads fail for files > 8MB (default threshold)
- **Go SDK**: Multipart uploads fail, list objects with delimiters returns no prefixes
- **Node.js SDK**: Range-based downloads fail, conditional requests ignored
- **Java SDK**: Generation-based operations broken, checksums missing

## What Changes

- **List Objects with commonPrefixes**: Implement delimiter-based grouping so SDKs can navigate "directory" structures
- **Range Requests**: Full HTTP 206 Partial Content support with proper Content-Range headers
- **Generations & Versioning**: Switch from string to int64 generations, proper versioning lifecycle
- **Multipart Upload**: Parse multipart/form-data for metadata + content in single request
- **Resumable Upload Sessions**: Persistent session state with chunk accumulation
- **Checksums**: Auto-calculate CRC32C and MD5 on upload, return in response headers
- **Response Headers**: Full suite of X-Goog-* headers, ETag, Last-Modified, Accept-Ranges
- **Compose Objects**: Concatenate up to 32 source objects into one destination
- **Object Metadata**: ContentEncoding, ContentDisposition, ContentLanguage, CacheControl, CustomTime
- **Conditional Requests**: ifGenerationMatch, ifGenerationNotMatch, ifMetagenerationMatch, ifMetagenerationNotMatch

## Capabilities

### Modified Capabilities

- `gcs-api`: Extended to include commonPrefixes, range requests, conditional requests, compose, full metadata, response headers
- `storage-backend`: Extended to support int64 generations, versioning, chunk accumulation for resumable uploads, checksum calculation

## Impact

- **internal/backend**: Backend interface changes (generation type, new methods for chunked uploads)
- **internal/api**: Handler changes for multipart parsing, range headers, conditional logic
- **internal/model**: Object struct additions for metadata fields
- **internal/router**: Route additions for compose, conditional request middleware
- **Compatibility**: After this change, Go/Python/Node.js SDKs should work for core operations
