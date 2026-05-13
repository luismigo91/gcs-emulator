# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /gcs-emulator ./cmd/server

# Final stage
FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /

COPY --from=builder /gcs-emulator /gcs-emulator

EXPOSE 9090

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:9090/-/health || exit 1

ENV GCP_EMULATOR_PUBSUB_MODE=memory

ENTRYPOINT ["/gcs-emulator"]
