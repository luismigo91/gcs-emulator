## ADDED Requirements

### Requirement: GCS object events deliver Pub/Sub messages
When a configured notification event occurs on a GCS bucket, the emulator SHALL publish a Pub/Sub message to the corresponding topic, enabling end-to-end event-driven testing.

#### Scenario: OBJECT_FINALIZE triggers publish
- **GIVEN** a bucket has a notification config for topic `my-topic` with event type `OBJECT_FINALIZE`
- **AND** a subscription exists for that topic
- **WHEN** an object is created in the bucket
- **THEN** a Pub/Sub message is published to the topic with attributes `bucketId`, `objectId`, `eventType`, and `notificationConfig`

#### Scenario: OBJECT_DELETE triggers publish
- **GIVEN** a notification config with event type `OBJECT_DELETE`
- **WHEN** an object is deleted
- **THEN** a Pub/Sub message is delivered to subscribers

#### Scenario: No notification config means no publish
- **GIVEN** a bucket with no notification configuration
- **WHEN** an object is created
- **THEN** no Pub/Sub message is published

#### Scenario: Multiple notification configs per bucket
- **GIVEN** a bucket has 2 notification configs for different topics
- **WHEN** an object is created
- **THEN** messages are published to both topics
