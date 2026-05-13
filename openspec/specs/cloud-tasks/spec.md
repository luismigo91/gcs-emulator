# cloud-tasks

## Requirements

### Requirement: Queue CRUD
The emulator SHALL support creating, retrieving, listing, updating, and deleting task queues.

#### Scenario: Create and list queues
- **WHEN** POST /v2/projects/{p}/locations/{l}/queues?queueId=my-q
- **THEN** queue is created with state RUNNING

### Requirement: Queue Lifecycle
The emulator SHALL support pausing, resuming, and purging queues.

#### Scenario: Pause and resume
- **WHEN** POST .../queues/{q}:pause
- **THEN** queue state is PAUSED; :resume returns to RUNNING

### Requirement: Task CRUD
The emulator SHALL support creating, retrieving, listing, and deleting tasks.

#### Scenario: Create task with HTTP target
- **WHEN** POST .../queues/{q}/tasks with httpRequest
- **THEN** task is created and auto-dispatched if queue is RUNNING

### Requirement: Task Execution
The emulator SHALL dispatch tasks to their HTTP targets when created or run manually.

#### Scenario: Manual run
- **WHEN** POST .../tasks/{t}:run
- **THEN** the HTTP request is executed against the target
