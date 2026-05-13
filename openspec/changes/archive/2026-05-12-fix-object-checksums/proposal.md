## Why

Objects uploaded to the emulator have empty `crc32c` and `md5Hash` fields, and the `X-Goog-Hash` response header returns `crc32c=,md5=`. The checksum calculation utilities (`util.CalculateCRC32C`, `util.CalculateMD5`) exist and are tested but are never called during object creation. This breaks integrity verification for any client that validates checksums, including the official Go SDK which does this by default.

## What Changes

- `MemoryBackend.CreateObject()` will calculate CRC32C and MD5 checksums when storing object content
- `Object` model instances will have populated `CRC32C` and `MD5Hash` fields after creation
- `X-Goog-Hash` response headers will return actual checksum values instead of empty strings
- HEAD and GET requests will return correct checksum headers

## Capabilities

### New Capabilities
<!-- None - this is a bug fix for an existing capability -->

### Modified Capabilities
- `gcs-api`: The "Checksums" requirement already exists in the spec but was not implemented. The implementation will now satisfy the existing requirement scenarios.

## Impact

- `internal/backend/memory.go`: Add checksum calculation to `CreateObject()`
- `internal/backend/seed.go`: Add checksum calculation to `seedBucket()`
- All object creation paths will now populate checksums automatically
- No breaking changes - empty checksums become non-empty, which is strictly additive
