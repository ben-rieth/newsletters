.PHONY: dev

dev:
	@trap 'kill 0' SIGINT; (cd api && air) & (cd web && pnpm dev) & wait
