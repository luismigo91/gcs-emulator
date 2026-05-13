## MODIFIED Requirements

### Requirement: Checksums

The emulator SHALL calculate and return CRC32C and MD5 checksums for uploaded objects.

#### Scenario: Checksums calculated on upload

**WHEN** a client uploads an object with content "hello world"
**THEN** the object SHALL have:
- `crc32c`: base64-encoded CRC32C checksum using Google polynomial (0x82F63B78)
- `md5Hash`: base64-encoded MD5 hash

#### Scenario: Checksums returned in download headers

**WHEN** a client downloads an object
**THEN** the `X-Goog-Hash` header SHALL include both `crc32c=<value>` and `md5=<value>` with non-empty values
