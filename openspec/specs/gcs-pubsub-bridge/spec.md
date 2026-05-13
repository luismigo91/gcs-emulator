# gcs-pubsub-bridge

## Requirements

### Requirement: Object Events Trigger Pub/Sub
The emulator SHALL publish Pub/Sub messages when GCS object events match configured bucket notifications.

#### Scenario: OBJECT_FINALIZE triggers publish
- **GIVEN** bucket has notification config for Pub/Sub topic with event type OBJECT_FINALIZE
- **WHEN** an object is created in the bucket
- **THEN** a Pub/Sub message is published to the topic with attributes bucketId, objectId, eventType, and notificationConfig

#### Scenario: OBJECT_DELETE triggers publish
- **GIVEN** notification config with event type OBJECT_DELETE
- **WHEN** an object is deleted
- **THEN** a Pub/Sub message is delivered to all subscribers

#### Scenario: No notification config means no publish
- **GIVEN** bucket with no notification configuration
- **WHEN** objects are created or deleted
- **THEN** no Pub/Sub messages are published

#### Scenario: Multiple notification configs
- **GIVEN** bucket with 2 notification configs for different topics
- **WHEN** an object event occurs
- **THEN** messages are published to both topics

#### Scenario: Object name prefix filter
- **GIVEN** notification config with object_name_prefix "logs/"
- **WHEN** object "logs/jan.txt" is created
- **THEN** a message is published; creating "data.txt" does not trigger publish

#### Scenario: Subscribe and receive
- **GIVEN** a subscription on the notification topic
- **WHEN** a notification-triggered publish occurs
- **THEN** the subscriber can pull and acknowledge the notification message
