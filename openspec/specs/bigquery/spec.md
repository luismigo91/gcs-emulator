# bigquery

## Requirements

### Requirement: Dataset CRUD
The emulator SHALL support creating, retrieving, listing, and deleting datasets.

#### Scenario: Create dataset
- **WHEN** POST /bigquery/v2/projects/{p}/datasets with dataset reference
- **THEN** dataset is created

### Requirement: Table CRUD
The emulator SHALL support creating, retrieving, listing, and deleting tables with schemas.

#### Scenario: Create table with schema
- **WHEN** POST .../datasets/{ds}/tables with fields array
- **THEN** table is created with specified schema

### Requirement: Insert Rows
The emulator SHALL support inserting rows into tables.

#### Scenario: Insert data
- **WHEN** POST .../tables/{t}:insertAll with rows
- **THEN** rows are stored and queryable

### Requirement: SQL Query
The emulator SHALL support basic SQL SELECT queries with WHERE, LIMIT, and column projection.

#### Scenario: SELECT with WHERE and LIMIT
- **WHEN** POST /bigquery/v2/projects/{p}/queries with SELECT query
- **THEN** matching rows are returned with schema
