# GCS Emulator

A lightweight, high-performance Google Cloud Storage emulator written in Go. Inspired by [Floci](https://github.com/floci) for AWS.

## Overview

GCS Emulator provides a local development environment that mimics Google Cloud Storage, allowing you to develop and test cloud-dependent applications without needing actual GCP credentials or incurring costs.

## Features

- **Full GCS API emulation**: Both JSON API (`storage/v1`) and XML API
- **Multiple storage modes**: Memory, Persistent, Hybrid (async flush), and WAL (write-ahead log)
- **Multi-project isolation**: Resources scoped by `GOOGLE_CLOUD_PROJECT`
- **SDK compatible**: Works with official Google Cloud SDKs (Go, Python, Node.js, Java)
- **Docker-first**: Single Docker image (~15-20MB) with `docker compose` support
- **Go library API**: Importable package for programmatic server control in tests
- **No credentials required**: Accepts any credentials, perfect for local development

## Quick Start

### Using Docker

```bash
docker-compose up -d
```

The emulator will be available at `http://localhost:9090`.

### Using Go

```bash
go run ./cmd/server
```

### Using Make

```bash
make run
```

## SDK Usage Examples

### Go

```go
import (
    "cloud.google.com/go/storage"
    "google.golang.org/api/option"
)

client, err := storage.NewClient(ctx,
    option.WithEndpoint("http://localhost:9090/storage/v1/"),
    option.WithoutAuthentication(),
)
```

### Python

```python
from google.cloud import storage

client = storage.Client(
    project="test-project",
    client_options={"api_endpoint": "http://localhost:9090"}
)
```

### Node.js

```javascript
const { Storage } = require('@google-cloud/storage');

const storage = new Storage({
  apiEndpoint: 'http://localhost:9090',
  projectId: 'test-project',
  credentials: {
    client_email: 'test@test.iam.gserviceaccount.com',
    private_key: 'test',
  },
});
```

### Java

```java
Storage storage = StorageOptions.newBuilder()
    .setHost("http://localhost:9090")
    .setProjectId("test-project")
    .build()
    .getService();
```

## Configuration

All configuration is via environment variables with `GCP_EMULATOR_` prefix:

| Variable | Default | Description |
|----------|---------|-------------|
| `GCP_EMULATOR_PORT` | `9090` | Port to listen on |
| `GCP_EMULATOR_DEFAULT_PROJECT` | `test-project` | Default project ID |
| `GCP_EMULATOR_STORAGE_MODE` | `memory` | Storage mode: `memory`, `persistent`, `hybrid`, `wal` |
| `GCP_EMULATOR_STORAGE_PATH` | `./data` | Path for persistent storage |
| `GCP_EMULATOR_FLUSH_INTERVAL` | `5s` | Flush interval for hybrid mode |

### Per-Service Overrides

You can override storage mode per service:

```bash
GCP_EMULATOR_SERVICE_GCS=mode:wal
GCP_EMULATOR_SERVICE_GCS=path:/data/gcs
```

## Project Structure

```
gcs-emulator/
├── cmd/server/          # CLI entry point
├── internal/
│   ├── api/             # HTTP handlers
│   ├── backend/         # Storage backends
│   ├── config/          # Configuration
│   ├── model/           # Domain models
│   └── router/          # HTTP router
├── pkg/emulator/        # Public Go library API
└── tests/               # Test files
```

## Storage Modes

- **Memory**: Ephemeral, fastest, data lost on restart
- **Persistent**: JSON files on disk, save/load on shutdown/startup
- **Hybrid**: In-memory with async flush to disk (default 5s interval)
- **WAL**: Write-ahead log with compaction for maximum durability

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development

```bash
# Build
make build

# Run tests
make test

# Run linter
make lint

# Build Docker image
make docker-build
```

## License

MIT
