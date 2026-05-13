## ADDED Requirements

### Requirement: Get Bucket CORS
The emulator SHALL expose `GET /storage/v1/b/{bucket}/cors` to retrieve the CORS configuration for a bucket.

#### Scenario: Get CORS for bucket with rules
- **WHEN** a bucket has CORS rules configured
- **THEN** `GET /storage/v1/b/{bucket}/cors` returns HTTP 200 with JSON body `{"kind": "storage#cors", "items": [...]}`

#### Scenario: Get CORS for bucket without rules
- **WHEN** a bucket has no CORS configuration
- **THEN** `GET /storage/v1/b/{bucket}/cors` returns HTTP 200 with JSON body `{"kind": "storage#cors", "items": []}` or an empty cors array

#### Scenario: Get CORS for non-existent bucket
- **WHEN** `GET /storage/v1/b/nonexistent/cors` is called
- **THEN** the response SHALL be HTTP 404

### Requirement: Set Bucket CORS
The emulator SHALL expose `PATCH /storage/v1/b/{bucket}/cors` to update the CORS configuration for a bucket.

#### Scenario: Set CORS rules
- **WHEN** `PATCH /storage/v1/b/{bucket}/cors` is called with body `[{"origin": ["*"], "method": ["GET"], "maxAgeSeconds": 3600}]`
- **THEN** the bucket's CORS is updated and returns HTTP 200
