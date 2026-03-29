.PHONY: dev

dev:
	make -C api dev & make -C web dev

generate:
	cd api && sqlc generate
	cd web && pnpm api:generate