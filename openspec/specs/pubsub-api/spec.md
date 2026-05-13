# pubsub-api

## Requirements

### Requirement: Topic CRUD
The emulator SHALL support Pub/Sub topic creation, retrieval, listing, and deletion.

#### Scenario: Create and list topics
- **WHEN** PUT /v1/projects/{p}/topics/{t} with topic name
- **THEN** topic is created and visible in GET list

### Requirement: Publish Messages
The emulator SHALL support publishing messages to a topic, delivering to all subscriptions.

#### Scenario: Publish delivers to subscriptions
- **WHEN** POST /v1/projects/{p}/topics/{t}:publish with messages
- **THEN** messageIds are returned and messages are queued for all matching subscriptions

### Requirement: Subscription CRUD
The emulator SHALL support subscription creation, retrieval, listing, and deletion.

#### Scenario: Create subscription on topic
- **WHEN** PUT /v1/projects/{p}/subscriptions/{s} with topic reference
- **THEN** subscription is created and visible

### Requirement: Pull Messages
The emulator SHALL support synchronous pull of messages from a subscription.

#### Scenario: Pull returns messages
- **WHEN** POST /v1/projects/{p}/subscriptions/{s}:pull with maxMessages
- **THEN** up to maxMessages are returned with ack IDs

### Requirement: Acknowledge Messages
The emulator SHALL support acknowledging messages after pull.

#### Scenario: Ack removes messages
- **WHEN** POST /v1/projects/{p}/subscriptions/{s}:acknowledge with ack IDs
- **THEN** those messages are removed from the subscription queue

### Requirement: Modify Ack Deadline
The emulator SHALL support extending ack deadlines.

#### Scenario: Extend deadline
- **WHEN** POST .../subscriptions/{s}:modifyAckDeadline with ack IDs and seconds
- **THEN** the deadline is updated

### Requirement: Schema CRUD
The emulator SHALL support schema creation, retrieval, listing, validation, commit, and deletion.

#### Scenario: Create Avro schema
- **WHEN** POST /v1/projects/{p}/schemas with type and definition
- **THEN** schema is created with revision

### Requirement: Snapshots
The emulator SHALL support snapshot creation, retrieval, and deletion.

#### Scenario: Create snapshot
- **WHEN** PUT /v1/projects/{p}/snapshots/{snap} with subscription
- **THEN** snapshot is created

### Requirement: Dead Letter Policy
The emulator SHALL route messages to dead letter topic when max delivery attempts are exceeded.

#### Scenario: DLQ delivery after max attempts
- **GIVEN** subscription with deadLetterPolicy maxDeliveryAttempts=2
- **WHEN** message delivery fails twice
- **THEN** message is published to the dead letter topic

### Requirement: Message Filtering
The emulator SHALL filter messages based on subscription filter criteria.

#### Scenario: Filter by attribute
- **GIVEN** subscription with filter "attributes.env = \"prod\""
- **WHEN** messages with env=prod and env=dev are published
- **THEN** only env=prod message is delivered
