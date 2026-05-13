## Context

The emulator has working checksum utilities (`util.CalculateCRC32C`, `util.CalculateMD5`) with passing tests, but these functions are never called during object creation. The `MemoryBackend.CreateObject()` method reads the content into a byte slice but only sets `Size`, `Generation`, and `Metageneration` on the object model. The `CRC32C` and `MD5Hash` fields remain empty strings.

The `X-Goog-Hash` header in response handlers formats these empty fields as `crc32c=,md5=`, which fails integrity checks in the official Go SDK.

## Goals / Non-Goals

**Goals:**
- Calculate CRC32C and MD5 for every object at creation time
- Populate `CRC32C` and `MD5Hash` fields on the `Object` model
- Ensure all object creation paths (simple upload, multipart, resumable, seed) produce checksums
- Pass existing SDK compatibility tests that verify checksum headers

**Non-Goals:**
- Client-provided checksum validation (rejecting uploads with mismatched checksums)
- MD5 vs CRC32C preference configuration
- gRPC checksum support

## Decisions

### Calculate checksums in `CreateObject()`, not in handlers

**Decision**: Add checksum calculation directly in `MemoryBackend.CreateObject()` after reading the content.

**Rationale**: This is the single point where all object content flows through. Whether the object comes from a simple upload, multipart upload, resumable upload finalization, or seed directory, they all call `CreateObject()`. Computing checksums here ensures consistency and avoids duplicating logic across multiple handlers.

**Alternative considered**: Calculate in each handler before calling `CreateObject()`. Rejected because it would require changes to 4+ handler methods and the seed module, with risk of missing a path.

### Use Google CRC32C polynomial (already implemented)

**Decision**: Continue using `crc32.MakeTable(0x82F63B78)` which is the GCS-specific polynomial.

**Rationale**: Already implemented and tested. The standard CRC32 polynomial would produce incorrect values for GCS clients.

### No checksum validation on upload

**Decision**: Do not validate client-provided `x-goog-hash` or `Content-MD5` headers on upload.

**Rationale**: This is a separate feature (integrity verification) that adds complexity. The immediate need is to return correct checksums so clients can verify them. Validation can be added later as a separate change.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Performance impact from computing checksums on every upload | CRC32C and MD5 are fast; impact is negligible for typical object sizes. For very large objects (>100MB), the SDK uploads in chunks anyway. |
| Seed directory scanning computes checksums for all files | Seed runs once at startup; acceptable one-time cost. |
| Existing tests that expect empty checksums may fail | Tests should be updated to expect non-empty values; this is the correct behavior. |
