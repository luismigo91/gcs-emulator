# cloud-logging

## Requirements

### Requirement: Write Log Entries
The emulator SHALL accept structured log entries via the Cloud Logging API.

#### Scenario: Write entries
- **WHEN** POST /v2/entries:write with log entries (logName, severity, textPayload/jsonPayload)
- **THEN** entries are stored with timestamps

### Requirement: List Log Entries
The emulator SHALL return stored log entries via the admin endpoint.

#### Scenario: List entries
- **WHEN** GET /-/logs with optional logName filter
- **THEN** most recent entries are returned in reverse chronological order
