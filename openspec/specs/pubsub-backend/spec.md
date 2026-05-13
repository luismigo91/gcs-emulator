# pubsub-backend

## Requirements

### Requirement: Memory mode
The emulator SHALL support ephemeral Pub/Sub message storage in memory.

#### Scenario: Data lost on restart
- **WHEN** emulator stops in memory mode
- **THEN** all topics, subscriptions, and messages are lost

### Requirement: Persistent mode
The emulator SHALL persist Pub/Sub topics, subscriptions, and schemas to disk.

#### Scenario: Data survives restart
- **WHEN** topics and subscriptions are created in persistent mode and emulator restarts
- **THEN** they are fully restored (messages are not persisted)

### Requirement: Hybrid mode
The emulator SHALL support async flushing of Pub/Sub state to disk.

#### Scenario: Flush on shutdown
- **WHEN** emulator shuts down in hybrid mode
- **THEN** a final flush saves topics, subscriptions, and schemas

### Requirement: WAL mode
The emulator SHALL support write-ahead logging for Pub/Sub operations.

#### Scenario: WAL replay on restart
- **WHEN** operations are logged to WAL before applying
- **THEN** on restart, WAL is replayed to restore state
