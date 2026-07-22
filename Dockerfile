# Build stage
FROM node:lts-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN corepack enable pnpm && pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm build

FROM golang:1.26.2-alpine AS builder

WORKDIR /app
COPY --from=frontend /app/web/dist ./web/dist
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -tags prod -ldflags="-s -w" -o api ./cmd/server

# Final stage
FROM debian:bookworm-slim AS dbmate
# Install dbmate
RUN apt-get update && apt-get install -y curl && \
    curl -fsSL -o /usr/local/bin/dbmate https://github.com/amacneil/dbmate/releases/download/v2.32.0/dbmate-linux-amd64 && \
    chmod +x /usr/local/bin/dbmate && rm -rf /var/lib/apt/lists/*

FROM gcr.io/distroless/static-debian12    
# Copy binary and migrations from builder
COPY --from=builder /app/api .
COPY --from=builder /app/db/migrations ./db/migrations
COPY --from=dbmate /usr/local/bin/dbmate /usr/local/bin/dbmate

CMD ["./api"]