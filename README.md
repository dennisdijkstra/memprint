# WIP:Memprint

Upload a file. Get back a poster made from its memory fingerprint.

Every upload captures real runtime data from the Go process — heap addresses,
PIDs, syscall numbers, file descriptors — and renders them into a unique
typographic poster. No two posters are the same.

## Tech Stack

- Go, Node.js (TypeScript)
- gRPC + Protobuf
- RabbitMQ
- PostgreSQL
- Redis
- AWS S3
- Docker

## Architecture
```
POST /upload
     │
     ▼ gRPC
API Gateway (:8080)
     │
     └── File Service (:50051)
               ├── captures PID, heap addr, syscalls, fd
               ├── stores to PostgreSQL
               └── publishes → file.uploaded (RabbitMQ)
                                    │
                                    ▼
                            Render Service (Go)
                               ├── builds layout from metadata
                               ├── calls → Node Renderer (:50053) gRPC
                               │                └── renders poster via canvas
                               ├── uploads PNG to S3
                               └── publishes → poster.ready (RabbitMQ)
                                                    │
                                                    ▼
                                        Notification Service
                                            └── sends email via Resend
```

## Project Structure
```
.
├── docker-compose.yml
├── go.mod
├── go.sum
├── proto/
│   ├── file.proto
│   └── file/                  # generated protobuf + gRPC stubs
├── services/
│   ├── gateway/               # HTTP API gateway, rate limiting
│   ├── file/                  # file upload, metadata capture
│   ├── render/                # layout engine, delegates rendering via gRPC
│   ├── renderer/              # Node.js gRPC service, renders poster via canvas
│   └── notifications/         # email delivery via Resend
└── shared/
    └── events/                # shared queue names + event types
```

## Getting Started

1. Install Go 1.25+
2. Download dependencies:
```bash
go mod download
```

3. Copy `.env.example` to `.env` and fill in the values:
```bash
cp .env.example .env
```

4. Start everything:
```bash
docker compose up
```

5. Upload a file:
```bash
curl -X POST http://localhost:8080/upload \
  -F "user_id=user_123" \
  -F "file=@yourfile.png"
```

You'll receive an email with a link to your generated poster.

## Development

Run all tests:
```bash
go test ./...
```

Run services individually (outside Docker):
```bash
go run ./services/file/
go run ./services/gateway/
go run ./services/render/
go run ./services/notifications/
cd services/renderer && npm run dev
```

Test the gRPC endpoint directly with grpcurl:
```bash
grpcurl -plaintext \
  -d '{"user_id":"user_123","filename":"hello.txt","content":"aGVsbG8="}' \
  127.0.0.1:50051 file.FileService/UploadFile
```
