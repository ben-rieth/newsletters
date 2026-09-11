# dash's job control is unreliable in non-interactive scripts; `dev` needs `set -m`
SHELL := /bin/bash

.PHONY: dev build build-web build-server

.DEFAULT_GOAL := build-server

# set -m puts each child in its own process group, so Ctrl-C reaches only this
# shell and the trap can tear things down in order. Without it the terminal
# SIGINTs dlv and the debuggee simultaneously, which leaves an orphan on :8080
# and can segfault the ptrace-stopped binary.
dev:
	@set -m; \
	(cd web && pnpm dev) & web=$$!; \
	air & air=$$!; \
	trap 'kill -TERM -$$web 2>/dev/null; kill -TERM -$$air 2>/dev/null; \
	      sleep 3; \
	      kill -KILL -$$web 2>/dev/null; kill -KILL -$$air 2>/dev/null; \
	      fuser -k 8080/tcp 2345/tcp >/dev/null 2>&1; \
	      exit 0' INT TERM; \
	wait $$air

build: build-web build-server

build-web:
	cd web && npm run build

# web/dist is gitignored; //go:embed all:dist needs it to exist with at least one file
build-server:
	@mkdir -p web/dist && touch web/dist/.gitkeep
	go build -o bin/server ./cmd/server
