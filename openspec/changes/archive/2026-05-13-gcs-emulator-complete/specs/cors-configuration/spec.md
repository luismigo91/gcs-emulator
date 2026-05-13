## ADDED Requirements

### Requirement: CORS Configuration

The emulator SHALL support bucket CORS rule configuration storage and reporting.

#### Scenario: Set CORS rules

**WHEN** a client sends `PATCH /storage/v1/b/{bucket}` with CORS configuration
**THEN** the CORS rules SHALL be stored with the bucket

#### Scenario: Get CORS rules

**WHEN** a client retrieves bucket metadata
**THEN** the response SHALL include the configured CORS rules
