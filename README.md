# WIP: Memprint Project

Upload a file. Get back a poster made from its memory fingerprint.

## Tech Stack

- Go
- gRPC + Protobuf
- HTTP gateway
- PostgreSQL, RabbitMQ, Redis (via Docker Compose)

## Project Structure

```text
.
|- docker-compose.yml
|- go.mod
|- go.sum
|- proto/
|  |- file.proto
|  `- file/                  # generated protobuf and gRPC code
|- services/
|  |- file/                  # file upload gRPC service
|  |- gateway/               # HTTP API gateway
|  `- render/                # poster rendering service
`- shared/
	`- events/                # shared event definitions
```

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

4. Copy `.env.example` to `.env`, fill in the values, and run services:

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