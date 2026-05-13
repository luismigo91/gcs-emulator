# cloud-cdn

## Requirements

### Requirement: Backend Service CRUD
The emulator SHALL support creating, retrieving, listing, and deleting CDN backend services.

#### Scenario: Create backend with CDN enabled
- **WHEN** POST /compute/v1/projects/{p}/global/backendServices with enableCDN=true
- **THEN** backend service is created

### Requirement: URL Map CRUD
The emulator SHALL support creating, retrieving, listing, and deleting URL maps for CDN routing.

#### Scenario: Create URL map with host rules
- **WHEN** POST /compute/v1/projects/{p}/global/urlMaps with default service
- **THEN** URL map is created
