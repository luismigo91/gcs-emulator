# Uso con Terraform

El emulador expone un endpoint de autenticación OAuth2 y las APIs de GCP, lo que permite usar **cualquier provider de Terraform** que soporte tokens estáticos y endpoints personalizados.

## Token de acceso

El emulador siempre devuelve el mismo token. No necesitas obtenerlo dinámicamente:

```
ya29.emulator.mock.access.token
```

Puedes verificarlo manualmente:

```bash
curl -s -X POST http://localhost:9090/oauth2/v4/token \
  -d "grant_type=client_credentials" | jq .
```

## Configuración del provider

```hcl
provider "google" {
  project      = "test-project"
  access_token = "ya29.emulator.mock.access.token"
  region       = "us-central1"
}

provider "google-beta" {
  project      = "test-project"
  access_token = "ya29.emulator.mock.access.token"
  region       = "us-central1"
}
```

## Variables de entorno

Los SDKs de Google (Go, Python, Node.js) respetan estas variables para redirigir el tráfico al emulador. Terraform los usa internamente:

```bash
export STORAGE_EMULATOR_HOST=http://localhost:9090
export PUBSUB_EMULATOR_HOST=http://localhost:9090
export GOOGLE_CLOUD_PROJECT=test-project
```

## Servicios soportados

| Recurso Terraform | Endpoint emulado | Soporte |
|-------------------|------------------|---------|
| `google_storage_bucket` | `/storage/v1/` | Buckets, objects, ACLs, IAM, lifecycle, CORS |
| `google_storage_bucket_object` | `/upload/storage/v1/` | Upload, download, compose, copy, rewrite |
| `google_pubsub_topic` | `/v1/projects/{p}/topics/` | Topics, schemas |
| `google_pubsub_subscription` | `/v1/projects/{p}/subscriptions/` | Push/pull, ack, dead letter, filtering |
| `google_secret_manager_secret` | `/v1/projects/{p}/secrets/` | Secrets CRUD, versions |
| `google_secret_manager_secret_version` | `/v1/projects/{p}/secrets/{s}/versions` | Access, lifecycle |
| `google_kms_key_ring` | `/v1/projects/{p}/locations/{l}/keyRings/` | Key rings |
| `google_kms_crypto_key` | `/v1/projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys` | Keys, versions, encrypt/decrypt |
| `google_cloud_tasks_queue` | `/v2/projects/{p}/locations/{l}/queues/` | Queues CRUD, tasks |
| `google_service_account` | `/v1/projects/{p}/serviceAccounts` | Service accounts, signJWT |
| `google_bigquery_dataset` | `/bigquery/v2/projects/{p}/datasets` | Datasets, tables, SQL SELECT |
| `google_bigquery_table` | `/bigquery/v2/projects/{p}/datasets/{d}/tables` | Tables |
| `google_cloud_scheduler_job` | `/v1/projects/{p}/locations/{l}/jobs` | HTTP + Pub/Sub targets |
| `google_dns_managed_zone` | `/dns/v1/projects/{p}/managedZones` | Managed zones, record sets |

## Servicios NO soportados

Cualquier recurso que no esté en la tabla anterior fallará. Esto incluye:
- Compute Engine (`google_compute_*`)
- Cloud Run (`google_cloud_run_*`)
- GKE (`google_container_*`)
- Cloud SQL (`google_sql_*`)
- VPC, Firewall, IAP, etc.

Para estos casos necesitarás un emulador diferente o usar GCP real.

## Limitaciones

1. **Auth simplificada**: el token es estático y nunca expira. No hay service accounts reales ni IAM granular.
2. **Sin persistencia por defecto**: en modo `memory` los datos se pierden al reiniciar. Usa `GCP_EMULATOR_STORAGE_MODE=persistent` para conservarlos.
3. **Sin operaciones regionales reales**: `location` y `region` son aceptados pero no validados.
4. **APIs parciales**: no todos los fields de cada recurso están implementados. Terraform puede enviar campos que el emulador ignora o rechaza.
5. **Sin soporte para gRPC**: solo HTTP/JSON. Algunos providers de Terraform usan gRPC internamente.
