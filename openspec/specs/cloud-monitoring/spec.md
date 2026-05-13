# cloud-monitoring

## Requirements

### Requirement: Create Time Series
The emulator SHALL accept time series data via the Monitoring API.

#### Scenario: Ingest time series
- **WHEN** POST /v3/.../timeSeries:createTimeSeries with metric data
- **THEN** time series points are stored

### Requirement: Query Time Series
The emulator SHALL support querying stored time series.

#### Scenario: List with filter
- **WHEN** GET /v3/.../timeSeries with filter on metric.type
- **THEN** matching time series are returned
