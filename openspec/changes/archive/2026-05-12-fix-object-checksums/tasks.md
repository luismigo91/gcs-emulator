## 1. Core Implementation

- [x] 1.1 Add checksum calculation to `MemoryBackend.CreateObject()` using `util.CalculateCRC32C` and `util.CalculateMD5`
- [x] 1.2 Add checksum calculation to `seed.go` `seedBucket()` function
- [x] 1.3 Verify `X-Goog-Hash` header returns non-empty values in HEAD and GET responses

## 2. Validation

- [x] 2.1 Run `go test -race ./...` and confirm all tests pass
- [x] 2.2 Run SDK compatibility tests and verify checksum headers are populated
- [x] 2.3 Manual test: upload object via curl and verify `X-Goog-Hash` header contains valid checksums
