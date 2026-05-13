## ADDED Requirements

### Requirement: Signed URL Support

The emulator SHALL support signed URL generation and validation for object access.

#### Scenario: Generate signed URL

**WHEN** a client requests a signed URL for an object
**THEN** the emulator SHALL return a URL with signature parameters

#### Scenario: Access object via signed URL

**WHEN** a client accesses an object using a signed URL
**THEN** the emulator SHALL serve the object content
