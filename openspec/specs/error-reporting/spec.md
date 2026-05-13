# error-reporting

## Requirements

### Requirement: Report Error Events
The emulator SHALL accept error reports via the Error Reporting API.

#### Scenario: Report error
- **WHEN** POST /v1beta1/projects/{p}/events:report with message and serviceContext
- **THEN** error event is stored

### Requirement: List Error Events
The emulator SHALL return stored error events.

#### Scenario: List errors
- **WHEN** GET /-/errors
- **THEN** most recent error events are returned
