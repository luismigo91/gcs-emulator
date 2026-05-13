# cloud-build

## Requirements

### Requirement: Trigger CRUD
The emulator SHALL support creating, retrieving, listing, and deleting build triggers.

#### Scenario: Create trigger
- **WHEN** POST /v1/projects/{p}/triggers with triggerTemplate or github config
- **THEN** trigger is created

### Requirement: Run Trigger
The emulator SHALL support manually running triggers to simulate builds.

#### Scenario: Run returns successful build
- **WHEN** POST .../triggers/{t}:run
- **THEN** build with status SUCCESS is returned
