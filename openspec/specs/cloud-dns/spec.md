# cloud-dns

## Requirements

### Requirement: Managed Zone CRUD
The emulator SHALL support creating, retrieving, listing, and deleting managed zones.

#### Scenario: Create zone
- **WHEN** POST /dns/v1/projects/{p}/managedZones?managedZoneId=my-zone
- **THEN** zone is created with nameservers

### Requirement: Record Set Management
The emulator SHALL support adding and listing resource record sets.

#### Scenario: Create change with record additions
- **WHEN** POST .../managedZones/{z}/changes with addition records
- **THEN** records are added and returned in rrsets list
