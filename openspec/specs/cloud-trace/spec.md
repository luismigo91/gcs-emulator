# cloud-trace

## Requirements

### Requirement: Batch Write Spans
The emulator SHALL accept trace spans via the Cloud Trace API.

#### Scenario: Write spans
- **WHEN** POST /v2/traces:batchWrite with spans (name, spanId, traceId, times)
- **THEN** spans are stored

### Requirement: List Traces
The emulator SHALL return stored spans.

#### Scenario: List all traces or get by traceId
- **WHEN** GET /-/traces with optional traceId query
- **THEN** matching spans are returned
