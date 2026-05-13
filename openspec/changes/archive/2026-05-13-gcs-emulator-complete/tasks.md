## 1. Technical Debt Fixes

- [x] 1.1 Fix ETag format: use base64-encoded generation instead of empty quotes
- [x] 1.2 Fix Content-Type auto-detection in all upload paths
- [x] 1.3 Fix ProjectNumber: derive from request context

## 2. IAM Policies

- [x] 2.1 Add IAM policy model types to `internal/model/`
- [x] 2.2 Implement getIamPolicy, setIamPolicy, testIamPermissions handlers
- [x] 2.3 Add IAM policy storage to MemoryBackend
- [x] 2.4 Add IAM policy routes to router
- [x] 2.5 Add integration tests for IAM policies

## 3. Bucket Notifications

- [x] 3.1 Add notification model types to `internal/model/`
- [x] 3.2 Implement notification CRUD handlers
- [x] 3.3 Add notification storage to MemoryBackend
- [x] 3.4 Add notification routes to router
- [x] 3.5 Add integration tests for notifications

## 4. Signed URLs

- [x] 4.1 Add signed URL generation utility to `internal/util/`
- [x] 4.2 Implement signed URL validation in handlers
- [x] 4.3 Add signed URL routes to router
- [x] 4.4 Add integration tests for signed URLs

## 5. Lifecycle Rules

- [x] 5.1 Add lifecycle rule model types to `internal/model/`
- [x] 5.2 Implement lifecycle CRUD handlers
- [x] 5.3 Add lifecycle storage to MemoryBackend
- [x] 5.4 Add lifecycle routes to router
- [x] 5.5 Add integration tests for lifecycle rules

## 6. CORS Configuration

- [x] 6.1 Add CORS model types to `internal/model/`
- [x] 6.2 Implement CORS CRUD handlers
- [x] 6.3 Add CORS storage to MemoryBackend
- [x] 6.4 Add CORS routes to router
- [x] 6.5 Add integration tests for CORS

## 7. SDK Compatibility

- [x] 7.1 Add Python SDK compatibility tests
- [x] 7.2 Add Node.js SDK compatibility tests
- [x] 7.3 Document SDK compatibility matrix

## 8. Production Readiness

- [x] 8.1 Add metrics package with Prometheus endpoint
- [x] 8.2 Add metrics middleware to router
- [x] 8.3 Add performance benchmark tests
- [x] 8.4 Optimize Dockerfile with multi-stage build

## 9. Validation

- [x] 9.1 Run `go test -race ./...` and confirm all tests pass
- [x] 9.2 Manual end-to-end test with curl
- [x] 9.3 Build and test Docker image
