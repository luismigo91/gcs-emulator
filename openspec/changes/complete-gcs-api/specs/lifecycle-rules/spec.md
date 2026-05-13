## ADDED Requirements

### Requirement: Get Bucket Lifecycle
The emulator SHALL expose `GET /storage/v1/b/{bucket}/lifecycle` to retrieve the lifecycle configuration for a bucket.

#### Scenario: Get lifecycle for bucket with rules
- **WHEN** a bucket has lifecycle rules configured
- **THEN** `GET /storage/v1/b/{bucket}/lifecycle` returns HTTP 200 with JSON body `{"kind": "storage#lifecycle", "rule": [...]}`

#### Scenario: Get lifecycle for bucket without rules
- **WHEN** a bucket has no lifecycle configuration
- **THEN** `GET /storage/v1/b/{bucket}/lifecycle` returns HTTP 200 with JSON body `{"kind": "storage#lifecycle", "rule": []}`

#### Scenario: Get lifecycle for non-existent bucket
- **WHEN** `GET /storage/v1/b/nonexistent/lifecycle` is called
- **THEN** the response SHALL be HTTP 404

### Requirement: Set Bucket Lifecycle
The emulator SHALL expose `PATCH /storage/v1/b/{bucket}/lifecycle` to update the lifecycle configuration for a bucket.

#### Scenario: Set lifecycle rules
- **WHEN** `PATCH /storage/v1/b/{bucket}/lifecycle` is called with body `{"rule": [{"action": {"type": "Delete"}, "condition": {"age": 30}}]}`
- **THEN** the bucket's lifecycle is updated and returns HTTP 200 with the updated lifecycle

### Requirement: Delete Bucket Lifecycle
The emulator SHALL expose `DELETE /storage/v1/b/{bucket}/lifecycle` to remove the lifecycle configuration from a bucket.

#### Scenario: Delete lifecycle
- **WHEN** `DELETE /storage/v1/b/{bucket}/lifecycle` is called
- **THEN** the bucket's lifecycle is removed and returns HTTP 204 No Content
