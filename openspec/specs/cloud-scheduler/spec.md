# cloud-scheduler

## Requirements

### Requirement: Job CRUD
The emulator SHALL support creating, retrieving, listing, updating, and deleting scheduled jobs.

#### Scenario: Create job with HTTP target
- **WHEN** POST /v1/projects/{p}/locations/{l}/jobs?jobId=my-job with httpTarget
- **THEN** job is created with state ENABLED

### Requirement: Job Lifecycle
The emulator SHALL support pausing, resuming, and manually running jobs.

#### Scenario: Run job manually
- **WHEN** POST .../jobs/{j}:run
- **THEN** the HTTP target or Pub/Sub publish is executed
