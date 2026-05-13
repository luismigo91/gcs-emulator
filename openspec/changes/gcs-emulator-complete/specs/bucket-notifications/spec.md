## ADDED Requirements

### Requirement: Bucket Notifications

The emulator SHALL support bucket notification configuration storage for SDK compatibility.

#### Scenario: Create notification

**WHEN** a client sends `POST /storage/v1/b/{bucket}/notificationConfigs` with notification config
**THEN** the notification SHALL be stored and returned with a generated ID

#### Scenario: List notifications

**WHEN** a client sends `GET /storage/v1/b/{bucket}/notificationConfigs`
**THEN** the response SHALL include all notification configurations for the bucket

#### Scenario: Get notification

**WHEN** a client sends `GET /storage/v1/b/{bucket}/notificationConfigs/{notification}`
**THEN** the response SHALL include the specified notification configuration

#### Scenario: Delete notification

**WHEN** a client sends `DELETE /storage/v1/b/{bucket}/notificationConfigs/{notification}`
**THEN** the notification SHALL be removed and response SHALL be HTTP 204
