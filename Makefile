.PHONY: dev build
dev:
	@trap 'fuser -k 8080/tcp 2345/tcp 2>/dev/null; kill 0' EXIT INT TERM; \
	(cd web && pnpm dev) & \
	air

build:
	cd web && npm run build
	go build -o bin/server ./cmd/server