## ADDED Requirements

### Requirement: Create Subscription
The emulator SHALL support `PUT /v1/projects/{project}/subscriptions/{subscription}` to create a Pub/Sub subscription attached to a topic.

#### Scenario: Create pull subscription
- **GIVEN** topic `my-topic` exists
- **WHEN** `PUT /v1/projects/test-project/subscriptions/my-sub` with `{"topic":"projects/test-project/topics/my-topic","ackDeadlineSeconds":10}`
- **THEN** the subscription is created and returned with HTTP 200

#### Scenario: Create subscription with missing topic
- **WHEN** creating a subscription referencing a non-existent topic
- **THEN** HTTP 404 is returned

### Requirement: Get Subscription
The emulator SHALL support `GET /v1/projects/{project}/subscriptions/{subscription}`.

#### Scenario: Get subscription
- **WHEN** `GET /v1/projects/test-project/subscriptions/my-sub` is called
- **THEN** the subscription resource is returned

### Requirement: List Subscriptions
The emulator SHALL support `GET /v1/projects/{project}/subscriptions`.

#### Scenario: List all subscriptions
- **WHEN** listing subscriptions for a project
- **THEN** all subscriptions are returned

### Requirement: Delete Subscription
The emulator SHALL support `DELETE /v1/projects/{project}/subscriptions/{subscription}`.

#### Scenario: Delete subscription
- **WHEN** `DELETE .../subscriptions/my-sub` is called
- **THEN** HTTP 200 is returned and subscription is removed

### Requirement: Pull Messages
The emulator SHALL support `POST /v1/projects/{project}/subscriptions/{subscription}:pull` to receive messages.

#### Scenario: Pull messages
- **GIVEN** messages have been published to the topic
- **WHEN** `POST ...:pull` with `{"maxMessages":10}` is called
- **THEN** up to 10 messages are returned with ack IDs

#### Scenario: Pull from empty subscription
- **GIVEN** no messages are pending
- **WHEN** pull is called
- **THEN** `{"receivedMessages":[]}` is returned (HTTP 200, not error)

### Requirement: Acknowledge Messages
The emulator SHALL support `POST /v1/projects/{project}/subscriptions/{subscription}:acknowledge` to ack messages.

#### Scenario: Ack messages
- **GIVEN** messages have been pulled with ack IDs
- **WHEN** `POST ...:acknowledge` with `{"ackIds":["ack-1","ack-2"]}` is called
- **THEN** those messages are removed and HTTP 200 is returned

### Requirement: Modify Ack Deadline
The emulator SHALL support `POST /v1/projects/{project}/subscriptions/{subscription}:modifyAckDeadline`.

#### Scenario: Extend ack deadline
- **WHEN** modifying ack deadline with `{"ackIds":["ack-1"],"ackDeadlineSeconds":60}`
- **THEN** the deadline is updated and HTTP 200 is returned
