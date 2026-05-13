## ADDED Requirements

### Requirement: IAM Policy Management

The emulator SHALL support bucket-level IAM policy operations for compatibility with GCS SDKs.

#### Scenario: Get IAM policy

**WHEN** a client sends `GET /storage/v1/b/{bucket}/iam`
**THEN** the response SHALL include:
- `bindings`: array of role/member bindings
- `etag`: policy etag for optimistic concurrency

#### Scenario: Set IAM policy

**WHEN** a client sends `POST /storage/v1/b/{bucket}/iam` with a policy body
**THEN** the policy SHALL be stored and returned

#### Scenario: Test IAM permissions

**WHEN** a client sends `POST /storage/v1/b/{bucket}/iam:testIamPermissions` with a permissions list
**THEN** the response SHALL include the subset of permissions that are allowed
