FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o render ./services/render/...

FROM alpine:3.20
RUN apk add --no-cache ca-certificates ttf-freefont
WORKDIR /app
COPY --from=builder /app/render .
COPY services/render/fonts/ ./fonts/
ENTRYPOINT ["./render"]
