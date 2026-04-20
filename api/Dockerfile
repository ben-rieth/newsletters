# Build stage
FROM golang:1.26 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o api ./cmd/server

# Final stage
FROM debian:bookworm-slim

WORKDIR /app

# Install dbmate
RUN apt-get update && apt-get install -y curl && \
    curl -fsSL -o /usr/local/bin/dbmate https://github.com/amacneil/dbmate/releases/download/v2.32.0/dbmate-linux-amd64 && \
    chmod +x /usr/local/bin/dbmate && rm -rf /var/lib/apt/lists/*

# Copy binary and migrations from builder
COPY --from=builder /app/api .
COPY --from=builder /app/db/migrations ./db/migrations

CMD ["./api"]