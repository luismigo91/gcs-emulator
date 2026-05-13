## ADDED Requirements

### Requirement: Node.js SDK bucket operations
The Node.js SDK test suite SHALL verify basic bucket CRUD operations against the running emulator using `@google-cloud/storage`.

#### Scenario: Create and list buckets via Node.js SDK
- **GIVEN** the emulator is running
- **WHEN** a Node.js script creates a bucket and lists buckets
- **THEN** the created bucket appears in the list

#### Scenario: Upload and download object via Node.js SDK
- **GIVEN** a bucket exists
- **WHEN** a Node.js script uploads an object and downloads it
- **THEN** the downloaded content matches the uploaded content

### Requirement: Test isolation
The Node.js SDK tests SHALL be self-contained and clean up created resources.

#### Scenario: Test cleanup
- **WHEN** a Node.js SDK test completes
- **THEN** any buckets or objects created during the test are deleted
