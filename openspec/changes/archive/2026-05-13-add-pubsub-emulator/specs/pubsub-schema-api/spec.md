## ADDED Requirements

### Requirement: Create Schema
The emulator SHALL expose POST /v1/projects/{project}/schemas to create a Pub/Sub schema.

#### Scenario: Create Avro schema
- **WHEN** POST /v1/projects/test-project/schemas is called with schema definition
- **THEN** HTTP 200 with the schema resource

### Requirement: Validate Schema
The emulator SHALL expose POST /v1/projects/{project}/schemas:validate.

#### Scenario: Validate valid schema
- **WHEN** POST .../schemas:validate is called with a valid definition
- **THEN** HTTP 200

### Requirement: Get Schema
The emulator SHALL expose GET /v1/projects/{project}/schemas/{schema}.

#### Scenario: Get schema
- **WHEN** GET .../schemas/my-schema is called
- **THEN** HTTP 200 with the schema resource including revision

### Requirement: List Schemas
The emulator SHALL expose GET /v1/projects/{project}/schemas.

#### Scenario: List schemas
- **WHEN** GET .../schemas is called
- **THEN** HTTP 200 with schemas array

### Requirement: Delete Schema
The emulator SHALL expose DELETE /v1/projects/{project}/schemas/{schema}.

#### Scenario: Delete schema
- **WHEN** DELETE .../schemas/my-schema is called
- **THEN** HTTP 200 and schema is removed
