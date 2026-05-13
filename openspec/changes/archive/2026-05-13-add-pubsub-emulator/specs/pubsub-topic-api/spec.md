## ADDED Requirements

### Requirement: Create Topic
The emulator SHALL support `PUT /v1/projects/{project}/topics/{topic}` to create a Pub/Sub topic.

#### Scenario: Create topic
- **WHEN** `PUT /v1/projects/test-project/topics/my-topic` is called with body `{"name":"projects/test-project/topics/my-topic"}`
- **THEN** the topic is created and returned with HTTP 200

#### Scenario: Duplicate topic returns conflict
- **GIVEN** topic `my-topic` exists
- **WHEN** creating `my-topic` again
- **THEN** HTTP 409 is returned

### Requirement: Get Topic
The emulator SHALL support `GET /v1/projects/{project}/topics/{topic}`.

#### Scenario: Get existing topic
- **WHEN** `GET /v1/projects/test-project/topics/my-topic` is called
- **THEN** the topic resource is returned with HTTP 200

#### Scenario: Get missing topic
- **WHEN** `GET /v1/projects/test-project/topics/nonexistent` is called
- **THEN** HTTP 404 is returned

### Requirement: List Topics
The emulator SHALL support `GET /v1/projects/{project}/topics` with pagination.

#### Scenario: List topics
- **GIVEN** multiple topics exist in a project
- **WHEN** `GET /v1/projects/test-project/topics` is called
- **THEN** all topics are returned with `{"topics": [...]}` format

#### Scenario: List with pageSize
- **WHEN** `GET /v1/projects/test-project/topics?pageSize=10` is called
- **THEN** at most 10 topics are returned with a `nextPageToken` if needed

### Requirement: Delete Topic
The emulator SHALL support `DELETE /v1/projects/{project}/topics/{topic}`.

#### Scenario: Delete topic
- **WHEN** `DELETE /v1/projects/test-project/topics/my-topic` is called
- **THEN** HTTP 200 is returned and the topic is removed

### Requirement: Publish Message
The emulator SHALL support `POST /v1/projects/{project}/topics/{topic}:publish` to publish messages to a topic.

#### Scenario: Publish single message
- **GIVEN** topic `my-topic` exists with subscription `my-sub`
- **WHEN** `POST .../topics/my-topic:publish` with `{"messages":[{"data":"aGVsbG8=","attributes":{"key":"value"}}]}`
- **THEN** HTTP 200 is returned with `{"messageIds":["1"]}` and the message is delivered to all subscriptions

#### Scenario: Publish to non-existent topic
- **WHEN** publishing to a topic that doesn't exist
- **THEN** HTTP 404 is returned
