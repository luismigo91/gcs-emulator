## ADDED Requirements

### Requirement: Lifecycle Rules

The emulator SHALL support bucket lifecycle rule configuration storage and reporting.

#### Scenario: Set lifecycle rules

**WHEN** a client sends `PATCH /storage/v1/b/{bucket}` with lifecycle configuration
**THEN** the lifecycle rules SHALL be stored with the bucket

#### Scenario: Get lifecycle rules

**WHEN** a client retrieves bucket metadata
**THEN** the response SHALL include the configured lifecycle rules
