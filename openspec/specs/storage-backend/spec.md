# storage-backend

## MODIFIED Requirements

### Requirement: IAM and Notification persistence
All disk-backed modes SHALL persist IAM policies, notification configurations, and ACLs alongside bucket/object data.

#### Scenario: IAM policy survives restart
- **WHEN** an IAM policy is set in persistent mode and the emulator restarts
- **THEN** the policy is restored and accessible

#### Scenario: Notifications survive restart
- **WHEN** notification configs are created and the emulator restarts
- **THEN** they are restored and listed correctly

### Requirement: Binary content persistence
Disk-backed modes SHALL persist object binary content as blob files alongside metadata JSON.

#### Scenario: Object content survives restart
- **WHEN** an object is created with content in persistent mode and emulator restarts
- **THEN** downloading the object returns the original content

### Requirement: ACL persistence
Disk-backed modes SHALL persist bucket, object, and default object ACLs.

#### Scenario: ACLs survive restart
- **WHEN** ACLs are set and emulator restarts
- **THEN** they are fully restored
