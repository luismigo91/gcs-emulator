## MODIFIED Requirements

### Requirement: Notification delivery
The emulator SHALL deliver Pub/Sub messages to configured notification topics when GCS object events occur.

#### Scenario: Notification storage unchanged
- **WHEN** a notification configuration is created via the GCS API
- **THEN** the configuration is stored as before (no API change)

#### Scenario: Actual message delivery
- **WHEN** a GCS object event matches a notification configuration
- **THEN** a Pub/Sub message is published to the configured topic in-process, and subscribers pulling from that topic receive the message

#### Scenario: Delivery failure is logged
- **WHEN** a notification-triggered publish fails (e.g., topic doesn't exist)
- **THEN** the failure is logged but does not block the GCS operation
