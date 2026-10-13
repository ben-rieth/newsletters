# syntax=docker/dockerfile:1.7
FROM node:24-alpine AS frontend
WORKDIR /app/web
RUN corepack enable pnpm
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN --mount=type=cache,id=s/9d699ec3-e8e1-4f80-84bf-05d20bfaed66-/pnpm/store,target=/pnpm/store \
    pnpm install --frozen-lockfile --store-dir /pnpm/store
COPY web/ .
RUN pnpm build

FROM golang:1.27.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,id=s/9d699ec3-e8e1-4f80-84bf-05d20bfaed66-/go/pkg/mod,target=/go/pkg/mod \
    go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY web/embed.go ./web/
COPY --from=frontend /app/web/dist ./web/dist
RUN --mount=type=cache,id=s/9d699ec3-e8e1-4f80-84bf-05d20bfaed66-/go/pkg/mod,target=/go/pkg/mod \
    --mount=type=cache,id=s/9d699ec3-e8e1-4f80-84bf-05d20bfaed66-/root/.cache/go-build,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -tags prod -ldflags="-s -w" -o api ./cmd/server

FROM gcr.io/distroless/static-debian12
ADD --chmod=755 https://github.com/amacneil/dbmate/releases/download/v2.32.0/dbmate-linux-amd64 /usr/local/bin/dbmate
COPY db/migrations ./db/migrations
COPY --from=builder /app/api .
CMD ["./api"]
