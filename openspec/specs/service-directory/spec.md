# service-directory

## Requirements

### Requirement: Namespace CRUD
The emulator SHALL support creating, retrieving, listing, and deleting namespaces.

#### Scenario: Create namespace
- **WHEN** POST /v1/projects/{p}/locations/{l}/namespaces?namespaceId=my-ns
- **THEN** namespace is created

### Requirement: Service CRUD
The emulator SHALL support creating and listing services within namespaces.

#### Scenario: Create service with endpoints
- **WHEN** POST .../namespaces/{ns}/services with endpoint address and port
- **THEN** service is created with endpoints
