## ADDED Requirements

### Requirement: Project-scoped resources
The emulator SHALL scope all GCS resources (buckets, objects, notifications) to a Google Cloud project ID. Resources created under one project SHALL NOT be visible to other projects.

#### Scenario: Create bucket in project A
- **WHEN** a bucket is created with project context `project-a`
- **THEN** the bucket is only visible when listing buckets under `project-a`

#### Scenario: Bucket isolation between projects
- **WHEN** project A creates a bucket named `my-bucket` and project B lists its buckets
- **THEN** project B's bucket list does not include `my-bucket`

#### Scenario: Same bucket name in different projects
- **WHEN** project A creates `my-bucket` and project B also creates `my-bucket`
- **THEN** both buckets exist independently with separate objects and metadata

### Requirement: Project ID resolution
The emulator SHALL resolve the project ID from request context using a defined priority order.

#### Scenario: X-Goog-User-Project header
- **WHEN** a request includes the `X-Goog-User-Project: my-project` header
- **THEN** the emulator uses `my-project` as the project ID for the operation

#### Scenario: Fallback to default project
- **WHEN** a request does not include any project identifier and `GCP_EMULATOR_DEFAULT_PROJECT` is set to `test-project`
- **THEN** the emulator uses `test-project` as the project ID

#### Scenario: Default project configuration
- **WHEN** the emulator starts with `GCP_EMULATOR_DEFAULT_PROJECT=dev-project`
- **THEN** all requests without explicit project context are scoped to `dev-project`

### Requirement: Project-aware API responses
The emulator SHALL include project context in API responses where the real GCP API does.

#### Scenario: Bucket list includes project
- **WHEN** listing buckets via `/storage/v1/b`
- **THEN** each bucket's `projectNumber` field reflects the owning project context

#### Scenario: Object metadata includes bucket
- **WHEN** retrieving object metadata
- **THEN** the `bucket` field in the response correctly identifies the bucket within its project scope
