# GCP Emulator

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/docker-~15MB-2496ED?logo=docker)](https://hub.docker.com)

A lightweight, zero-dependency **multi-service GCP emulator** for local development. Run GCS, Pub/Sub, Secret Manager, Cloud Tasks, and Cloud KMS — all in a single **15MB binary**, no credentials required.

## Services

| Service | API Prefix | Endpoints |
|---------|-----------|-----------|
| **Cloud Storage** | `/storage/v1/` | Buckets, objects, uploads, ACLs, IAM, lifecycle, CORS, notifications, compose, copy, rewrite |
| **Pub/Sub** | `/v1/projects/{p}/topics/` | Topics, subscriptions, publish, pull, ack, schemas, dead letter, message filtering |
| **Secret Manager** | `/v1/projects/{p}/secrets/` | Secrets CRUD, versions, access, lifecycle |
| **Cloud Tasks** | `/v2/projects/{p}/locations/{l}/queues/` | Queues CRUD, tasks, auto HTTP dispatch |
| **Cloud KMS** | `/v1/projects/{p}/locations/{l}/keyRings/` | Key rings, crypto keys, versions, encrypt/decrypt |

## Quick Start

```bash
# Homebrew
brew install luismiguelgilolivert/tap/gcs-emulator

# Docker
docker compose up -d

# Go install
go install github.com/luismiguelgilolivert/gcs-emulator/cmd/server@latest

# Go run
go run ./cmd/server
```

Server starts at `http://localhost:9090`.

## Usage Examples

### Go SDK

```go
// GCS
storage.NewClient(ctx, option.WithEndpoint("http://localhost:9090/storage/v1/"), option.WithoutAuthentication())

// Pub/Sub
pubsub.NewClient(ctx, "test-project", option.WithEndpoint("http://localhost:9090/v1/"), option.WithoutAuthentication())
```

### Python

```python
from google.cloud import storage, secretmanager

storage_client = storage.Client(project="test-project", client_options={"api_endpoint": "http://localhost:9090"})
secret_client = secretmanager.SecretManagerServiceClient(client_options={"api_endpoint": "http://localhost:9090"})
```

### Node.js

```javascript
const { Storage } = require('@google-cloud/storage');
const storage = new Storage({ apiEndpoint: 'http://localhost:9090', projectId: 'test-project' });
```

### curl

```bash
# Health check
curl http://localhost:9090/-/health

# Create bucket
curl -X POST http://localhost:9090/storage/v1/b -H "Content-Type: application/json" -d '{"name":"my-bucket"}'

# Upload object
curl -X POST "http://localhost:9090/upload/storage/v1/b/my-bucket/o?name=hello.txt" -d "Hello World"

# Create Pub/Sub topic
curl -X PUT http://localhost:9090/v1/projects/test-project/topics/demo-topic -d '{"name":"projects/test-project/topics/demo-topic"}'

# Create secret
curl -X POST "http://localhost:9090/v1/projects/test-project/secrets?secretId=my-secret" -d '{"replication":{"automatic":{}}}'
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GCP_EMULATOR_PORT` | `9090` | Listening port |
| `GCP_EMULATOR_DEFAULT_PROJECT` | `test-project` | Default project ID |
| `GCP_EMULATOR_STORAGE_MODE` | `memory` | `memory`, `persistent`, `hybrid`, `wal` |
| `GCP_EMULATOR_STORAGE_PATH` | `./data` | Path for persistent storage |
| `GCP_EMULATOR_FLUSH_INTERVAL` | `5s` | Flush interval for hybrid mode |
| `GCP_EMULATOR_PUBSUB_MODE` | `memory` | Pub/Sub storage mode |
| `GCP_EMULATOR_PUBSUB_PATH` | `./data/pubsub` | Pub/Sub storage path |
| `GCP_EMULATOR_SEED_PATH` | `/data` | Directory to preload on startup |

## Storage Modes

- **Memory**: Fastest, ephemeral
- **Persistent**: JSON + blobs on disk, survives restarts
- **Hybrid**: In-memory with async flush to disk
- **WAL**: Write-ahead log with compaction, maximum durability

## Architecture

```
gcs-emulator/
├── cmd/server/           # Binary entry point (5 services)
├── internal/
│   ├── api/              # GCS HTTP handlers
│   ├── backend/          # GCS storage backends
│   ├── auth/             # OAuth2 token endpoint
│   ├── pubsub/           # Pub/Sub (api, backend, model)
│   ├── secretmanager/    # Secret Manager
│   ├── cloudtasks/       # Cloud Tasks
│   ├── kms/              # Cloud KMS
│   ├── admin/            # /__/services, health
│   ├── metrics/          # Prometheus endpoint
│   ├── router/           # HTTP routing (all services)
│   └── config/           # Environment config
├── pkg/emulator/         # Programmatic server API
└── tests/                # Python + Node.js SDK tests
```

## Development

```bash
make build      # Build binary
make test       # Run tests with race detection
make lint       # Run linter
make docker-build  # Build Docker image
```

## License

MIT
