## ADDED Requirements

### Requirement: Prometheus Metrics

The emulator SHALL expose operational metrics in Prometheus format for monitoring.

#### Scenario: Metrics endpoint

**WHEN** a client sends `GET /metrics`
**THEN** the response SHALL include Prometheus-format metrics for:
- `gcs_emulator_requests_total`: total requests by method and path
- `gcs_emulator_request_duration_seconds`: request duration histogram
- `gcs_emulator_objects_total`: total objects across all buckets
- `gcs_emulator_buckets_total`: total buckets
