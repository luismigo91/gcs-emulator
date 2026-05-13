# docker-distribution

## MODIFIED Requirements

### Requirement: Docker image runs multiple services
The Docker image SHALL run all configured services (GCS, Pub/Sub, Secret Manager, Cloud Tasks, KMS, Logging, Monitoring, Error Reporting, Scheduler, IAM, Trace, BigQuery, DNS, Artifact Registry, Cloud Build, Billing, Asset Inventory, Service Directory, CDN).

#### Scenario: Docker container serves all services on port 9090
- **WHEN** docker run gcs-emulator is executed
- **THEN** all service APIs are available on port 9090

### Requirement: Health check reports all services
The health check endpoint SHALL report status of all running services.

#### Scenario: Health check multi-service
- **WHEN** GET /-/health is called
- **THEN** response includes all services with "available" status
