# Ejemplo de Terraform con GCS Emulator

## Requisitos

- Go 1.22+
- Terraform >= 1.0

## Uso

```bash
# 1. Arrancar el emulador
make run

# 2. En otra terminal, aplicar terraform
export STORAGE_EMULATOR_HOST=http://localhost:9090
export PUBSUB_EMULATOR_HOST=http://localhost:9090
export GOOGLE_CLOUD_PROJECT=test-project

terraform init
terraform plan
terraform apply -auto-approve
```

## Verificación

```bash
# Ver buckets
curl http://localhost:9090/storage/v1/b | jq .

# Ver topics de Pub/Sub
curl http://localhost:9090/v1/projects/test-project/topics | jq .

# Ver secrets
curl "http://localhost:9090/v1/projects/test-project/secrets" | jq .
```

## Limpiar

```bash
terraform destroy -auto-approve
```
