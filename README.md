# WIP: Memprint Project

Upload a file. Get back a poster made from its memory fingerprint.

## Tech Stack

- Go
- gRPC
- Protobuf
- Docker Compose (PostgreSQL, RabbitMQ, Redis)

## Project Structure

- `services/file/main.go` - gRPC file service server
- `services/file/db.go` - PostgreSQL connection helper
- `services/gateway/main.go` - HTTP API gateway
- `proto/file.proto` - protobuf service contract
- `proto/file/` - generated protobuf and gRPC Go files
- `docker-compose.yml` - local infrastructure services

## Getting Started

1. Install Go 1.25+.
2. Download dependencies:

```bash
go mod download
```

3. Start local infrastructure:

```bash
docker compose up -d
```

4. Copy the environment template, fill in values, and run services:

```bash
cp .env.example .env
go run services/file/*.go
go run services/gateway/*.go
```

## Development

- Run tests with `go test ./...`.
- Test the gRPC endpoint with grpcurl:

```bash
grpcurl -plaintext \
	-d '{"user_id":"user-123","filename":"hello.txt","content":"aGVsbG8="}' \
	127.0.0.1:50051 file.FileService/UploadFile
```