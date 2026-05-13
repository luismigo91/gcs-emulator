## Why

The GCS emulator is functionally complete for basic operations but has several gaps that prevent it from being a true drop-in replacement for GCS in production-like environments. This change addresses all remaining technical debt, missing API features, SDK compatibility gaps, and production readiness improvements in a single comprehensive update.

## What Changes

### Technical Debt
- **ETag format**: Replace empty quotes with actual hash values (generation-based)
- **Content-Type auto-detection**: Ensure all upload paths detect content type
- **ProjectNumber**: Derive from request context instead of always 0

### Missing GCS API Features
- **IAM policies**: getIamPolicy, setIamPolicy, testIamPermissions
- **Notifications**: create/list/delete bucket notifications
- **Signed URLs**: Generate and validate signed URLs for object access
- **Lifecycle rules**: Store and report lifecycle configuration
- **CORS configuration**: Store and return CORS rules for buckets

### SDK Compatibility
- **Python SDK tests**: Verify compatibility with google-cloud-storage
- **Node.js SDK tests**: Verify compatibility with @google-cloud/storage

### Production Readiness
- **Performance benchmarks**: Benchmark object operations at scale
- **Metrics endpoint**: Prometheus-compatible metrics endpoint
- **Docker optimization**: Multi-stage build for smaller image

## Capabilities

### New Capabilities
- `iam-policies`: Bucket-level IAM policy management
- `bucket-notifications`: Pub/Sub-style bucket notifications
- `signed-urls`: Signed URL generation and validation
- `lifecycle-rules`: Bucket lifecycle configuration
- `cors-configuration`: Bucket CORS rules management
- `metrics-endpoint`: Prometheus metrics for monitoring
- `sdk-python-compatibility`: Python SDK test coverage
- `sdk-nodejs-compatibility`: Node.js SDK test coverage
- `performance-benchmarks`: Performance testing infrastructure

### Modified Capabilities
- `gcs-api`: ETag, Content-Type, ProjectNumber fixes
- `docker-distribution`: Docker image optimization

## Impact

- `internal/model/`: New model types for IAM, notifications, lifecycle, CORS
- `internal/api/handler.go`: New handlers for IAM, notifications, signed URLs, lifecycle, CORS
- `internal/backend/memory.go`: Storage for new model types
- `internal/metrics/`: New metrics package
- `internal/util/`: ETag generation, signed URL utilities
- `pkg/emulator/`: New test files for Python/Node.js SDK
- `tests/benchmarks/`: New benchmark tests
- `Dockerfile`: Multi-stage build optimization
