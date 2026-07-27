.PHONY: dev build build-web build-server

.DEFAULT_GOAL := build-server

dev:
	@trap 'fuser -k 8080/tcp 2345/tcp 2>/dev/null; kill 0' EXIT INT TERM; \
	(cd web && pnpm dev) & \
	air

build: build-web build-server

build-web:
	cd web && npm run build

# web/dist is gitignored; //go:embed all:dist needs it to exist with at least one file
build-server:
	@mkdir -p web/dist && touch web/dist/.gitkeep
	go build -o bin/server ./cmd/server
