dev:
	cd web && pnpm dev &
	air

build:
	cd web && npm run build
	go build -o bin/server ./cmd/server