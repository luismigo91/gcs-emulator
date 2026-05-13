# secret-manager

## Requirements

### Requirement: Secret CRUD
The emulator SHALL support creating, retrieving, listing, and deleting secrets.

#### Scenario: Create and access secret
- **WHEN** POST /v1/projects/{p}/secrets?secretId=my-secret
- **THEN** secret is created and returned

### Requirement: Version Management
The emulator SHALL support adding, listing, accessing, enabling, disabling, and destroying versions.

#### Scenario: Add and access version
- **WHEN** POST /v1/projects/{p}/secrets/{s}:addVersion with payload
- **THEN** version is created and accessible via :access

#### Scenario: Disable and enable version
- **WHEN** version is disabled via :disable
- **THEN** it cannot be accessed; :enable restores access
