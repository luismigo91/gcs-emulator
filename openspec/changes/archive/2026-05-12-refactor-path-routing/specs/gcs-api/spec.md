## ADDED Requirements

### Requirement: Explicit Route Registration

The emulator SHALL register explicit URL route patterns for all API endpoints, ensuring consistent path parameter extraction without manual string parsing.

#### Scenario: Object operations route matching

**WHEN** a request matches `/storage/v1/b/{bucket}/o/{object...}`
**THEN** the `bucket` and `object` parameters SHALL be extracted correctly regardless of object name containing slashes

#### Scenario: Copy object route matching

**WHEN** a request matches `/storage/v1/b/{bucket}/o/{object}/copyTo/b/{destBucket}/o/{destObject...}`
**THEN** the source and destination bucket and object parameters SHALL be extracted correctly

#### Scenario: Rewrite object route matching

**WHEN** a request matches `/storage/v1/b/{bucket}/o/{object}/rewriteTo/b/{destBucket}/o/{destObject...}`
**THEN** the source and destination bucket and object parameters SHALL be extracted correctly
