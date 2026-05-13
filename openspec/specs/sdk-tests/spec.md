# sdk-tests

## Requirements

### Requirement: Go SDK Compatibility
The emulator SHALL be tested against the official Go Google Cloud SDKs.

#### Scenario: GCS SDK operations
- **WHEN** cloud.google.com/go/storage client is pointed at emulator
- **THEN** bucket and object CRUD operations succeed

### Requirement: Python SDK Compatibility
The emulator SHALL include Python SDK smoke tests for GCS.

#### Scenario: Python SDK creates bucket and uploads
- **WHEN** google-cloud-storage client connects to emulator
- **THEN** bucket CRUD, object upload/download, and list with prefix work

### Requirement: Node.js SDK Compatibility
The emulator SHALL include Node.js SDK smoke tests for GCS.

#### Scenario: Node SDK creates bucket and uploads
- **WHEN** @google-cloud/storage client connects to emulator
- **THEN** bucket CRUD, object upload/download, and list with prefix work

### Requirement: Test Cleanup
SDK tests SHALL clean up created resources after execution.

#### Scenario: Test isolation
- **WHEN** SDK tests complete
- **THEN** all created buckets, objects, topics, and subscriptions are deleted
