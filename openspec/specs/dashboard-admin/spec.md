# dashboard-admin

## Requirements

### Requirement: Web Dashboard
The emulator SHALL serve an HTML dashboard at /-/ showing running services and live statistics.

#### Scenario: Dashboard renders
- **WHEN** GET /-/ is called
- **THEN** HTML page is returned with service list, bucket count, topic count, and secret count

### Requirement: Live Statistics
The dashboard SHALL display real-time counts of emulated resources.

#### Scenario: Stats update
- **WHEN** resources are created or deleted
- **THEN** subsequent dashboard loads reflect the updated counts
