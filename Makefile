.PHONY: dev

dev:
	@trap 'kill 0' 2; (cd api && air) & (cd web && pnpm dev) & wait
