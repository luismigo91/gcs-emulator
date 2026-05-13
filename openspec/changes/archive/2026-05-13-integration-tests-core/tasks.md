## 1. HEAD Request Support

- [x] 1.1 Implement `HEAD /storage/v1/b/{bucket}` — Check bucket existence
- [x] 1.2 Implement `HEAD /storage/v1/b/{bucket}/o/{object}` — Check object existence
- [x] 1.3 Return correct headers (Content-Type, Content-Length, ETag, X-Goog-Generation) without body

## 2. Data Preloading

- [x] 2.1 Implement seed directory scanner: scan `{storagePath}/data/{bucket}/{object}` structure
- [x] 2.2 Create buckets from directory names
- [x] 2.3 Create objects from files with correct content type detection
- [x] 2.4 Integrate into `cmd/server/main.go` startup
- [x] 2.5 Add `SeedPath` option to `pkg/emulator` server constructor

## 3. Integration Tests — Bucket API

- [x] 3.1 Test `POST /storage/v1/b` — Create bucket with all fields
- [x] 3.2 Test `GET /storage/v1/b` — List buckets with project filter
- [x] 3.3 Test `GET /storage/v1/b/{bucket}` — Get bucket metadata
- [x] 3.4 Test `DELETE /storage/v1/b/{bucket}` — Delete empty bucket
- [x] 3.5 Test `DELETE /storage/v1/b/{bucket}` — Delete non-empty bucket (409)
- [x] 3.6 Test `PATCH /storage/v1/b/{bucket}` — Update bucket metadata
- [x] 3.7 Test `HEAD /storage/v1/b/{bucket}` — Check bucket existence

## 4. Integration Tests — Object API

- [x] 4.1 Test `POST /storage/v1/b/{bucket}/o` — Create object
- [x] 4.2 Test `GET /storage/v1/b/{bucket}/o` — List objects with prefix/delimiter
- [x] 4.3 Test `GET /storage/v1/b/{bucket}/o/{object}` — Get object metadata
- [x] 4.4 Test `GET /storage/v1/b/{bucket}/o/{object}?alt=media` — Download object
- [x] 4.5 Test `GET /storage/v1/b/{bucket}/o/{object}?alt=media` with Range header — Partial download
- [x] 4.6 Test `DELETE /storage/v1/b/{bucket}/o/{object}` — Delete object
- [x] 4.7 Test `PATCH /storage/v1/b/{bucket}/o/{object}` — Update object metadata
- [x] 4.8 Test `HEAD /storage/v1/b/{bucket}/o/{object}` — Check object existence
- [x] 4.9 Test `POST /storage/v1/b/{bucket}/o/{object}/compose` — Compose objects
- [x] 4.10 Test `POST /storage/v1/b/{bucket}/o/{object}/copyTo/...` — Copy object
- [x] 4.11 Test `POST /storage/v1/b/{bucket}/o/{object}/rewriteTo/...` — Rewrite object

## 5. Integration Tests — Uploads

- [x] 5.1 Test simple upload via `POST /upload/storage/v1/b/{bucket}/o`
- [x] 5.2 Test multipart upload with metadata
- [x] 5.3 Test resumable upload session creation
- [x] 5.4 Test resumable upload chunk upload
- [x] 5.5 Test resumable upload finalization

## 6. Integration Tests — XML API

- [x] 6.1 Test `PUT /{bucket}/{object}` — Upload object
- [x] 6.2 Test `GET /{bucket}/{object}` — Download object
- [x] 6.3 Test `GET /{bucket}` — List objects (XML format with CommonPrefixes)
- [x] 6.4 Test `DELETE /{bucket}/{object}` — Delete object
- [x] 6.5 Test `PUT /{bucket}` — Create bucket

## 7. Integration Tests — Versioning & Lifecycle

- [x] 7.1 Test versioning: overwrite preserves old generation
- [x] 7.2 Test versioning: list with versions=true shows all generations
- [x] 7.3 Test versioning: delete without generation preserves version
- [x] 7.4 Test generation-specific get

## 8. Concurrent Access Tests

- [x] 8.1 Test concurrent bucket creation (no duplicates)
- [x] 8.2 Test concurrent object writes (no data corruption)
- [x] 8.3 Test concurrent reads during writes (consistent reads)
- [x] 8.4 Run all tests with `go test -race`

## 9. Graceful Shutdown Tests

- [x] 9.1 Test persistent backend saves on shutdown
- [x] 9.2 Test hybrid backend flushes on shutdown
- [x] 9.3 Test WAL backend compacts on shutdown

## 10. Go SDK Compatibility Tests

- [x] 10.1 Set up test using `cloud.google.com/go/storage`
- [x] 10.2 Test bucket CRUD via Go SDK
- [x] 10.3 Test object CRUD via Go SDK
- [x] 10.4 Test object download/upload via Go SDK
- [x] 10.5 Test list objects with prefix/delimiter via Go SDK
- [x] 10.6 Document passing/failing endpoints
