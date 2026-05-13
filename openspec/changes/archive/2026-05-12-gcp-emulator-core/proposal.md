## Why

Google Cloud Platform lacks a unified, lightweight local emulator comparable to Floci for AWS. Existing solutions like `fake-gcs-server` only cover GCS in isolation, while LocalStack (the AWS equivalent) recently sunset its open-source community edition, creating a gap in the market for a free, multi-service GCP emulator. Developers currently must either test against real GCP (costly, slow, requires credentials) or piece together fragmented emulators that don't share a common architecture.

## What Changes

- **GCS Emulator (MVP)**: Full Google Cloud Storage API emulation supporting both JSON API and XML API, compatible with `google-cloud-go`, `google-cloud-python`, `@google-cloud/storage`, and `google-cloud-java` SDKs
- **Storage Backend System**: Four storage modes (memory, persistent, hybrid, WAL) configurable globally or per-service, inspired by Floci's architecture
- **Multi-Project Isolation**: Resources scoped by `GOOGLE_CLOUD_PROJECT`, enabling parallel development and testing across projects
- **Docker-First Distribution**: Single Docker image with `docker compose` support, environment variable configuration, and health endpoints
- **Go Library API**: Importable Go package for programmatic server control in tests
- **Plugin Architecture Foundation**: Service registration pattern that enables adding Pub/Sub, Firestore, Secret Manager, and other GCP services incrementally

## Capabilities

### New Capabilities

- `gcs-api`: Google Cloud Storage API emulation — bucket CRUD, object CRUD, resumable uploads, multipart uploads, downloads with range support, compose, copy, rewrite, versioning, lifecycle rules, IAM policies, ACLs, notifications, and both JSON API (`storage/v1`) and XML API endpoints
- `storage-backend`: Pluggable storage backend system with four modes — memory (ephemeral RAM), persistent (save/load on shutdown/startup), hybrid (RAM with async flush every 5s), and WAL (write-ahead log for maximum durability)
- `project-isolation`: Multi-project resource isolation where all GCS resources are scoped to a Google Cloud project ID, with configurable default project via environment variable
- `auth-emulation`: Fake authentication layer that accepts any credentials, supports service account JSON files, OAuth2 token endpoints, and Application Default Credentials (ADC) patterns without real validation
- `docker-distribution`: Docker image distribution with configurable ports via environment variables, health check endpoints, volume mounts for data persistence, and docker-compose ready configuration

### Modified Capabilities

<!-- No existing specs to modify — this is a greenfield project -->

## Impact

- **New repository**: Greenfield Go project with `cmd/`, `internal/`, `pkg/`, and `tests/` structure
- **External SDKs**: Must maintain wire-compatibility with official Google Cloud client libraries across Go, Python, Node.js, and Java
- **Docker ecosystem**: Produces a Docker image (~15-20MB Go binary) published to a container registry
- **OpenSpec**: First change in the project — establishes the spec-driven development workflow
- **README**: Updates the initial README with accurate project scope and getting-started instructions
