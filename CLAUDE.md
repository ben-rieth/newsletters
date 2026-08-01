# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Layout

This is a monorepo with a Go backend at the root and a React frontend in `web/`:

- `cmd/server/` — backend entry point
- `internal/` — backend packages (auth, config, db, email, feeds, handler, jobs, newsletters, templates, users, ...)
- `db/` — `migrations/` (timestamped SQL), `queries/` (sqlc source); sqlc generates Go into `internal/db/generated`
- `web/` — React frontend; `web/embed.go` embeds the built `web/dist` into the Go binary (`//go:embed all:dist`)

The frontend is served **embedded in the Go binary** in production — a single deployable artifact.

## Running

From the repo root:

```bash
make dev     # runs web (pnpm dev) + air (Go live reload) in parallel
make build   # builds web (npm run build) then compiles ./cmd/server into bin/server
```

- Backend: http://localhost:8080 (OpenAPI spec at `/openapi.json`)
- Frontend dev server: http://localhost:3001, proxying `/api` → `http://localhost:8080` (see `web/vite.config.ts`)

## Frontend Commands (run from `web/`)

```bash
pnpm dev          # Dev server on port 3001
pnpm build        # Production build → web/dist (embedded by the Go binary)
pnpm test         # Run Vitest tests (vitest run)
pnpm lint         # ESLint
pnpm check        # prettier --write + eslint --fix
pnpm api:generate # Regenerate TS types from backend OpenAPI spec (backend must be running)
```

Run a single test file / name: `pnpm exec vitest run src/path/to/file.test.ts -t "test name"`.

## Frontend Architecture (`web/`)

- **Framework:** [TanStack Start](https://tanstack.com/start) (React 19 meta-framework) with Vite.
- **Routing:** TanStack Router, file-based in `src/routes/`. The route tree (`routeTree.gen.ts`) is auto-generated — never edit it manually. Prefer directory routes over flat routes.
- **Feature modules:** `src/features/` is organized by domain (`auth`, `newsletters`, `issues`, `unsubscribe`, `debug`). Put feature-specific components, hooks, and logic here rather than in shared folders.
- **Import alias:** `#/*` maps to `src/*` (e.g. `import { cn } from '#/lib/utils'`).
- **Data fetching:** TanStack Query v5. `QueryClient` is injected into the router context; access it via `route.useRouteContext()`. Loaders prefetch with `queryClient.ensureQueryData()`, then components read with `useSuspenseQuery()`.
- **API client:** `openapi-fetch` in `src/api/client.ts`, base URL `/api`, `credentials: 'include'` (cookie-based auth). It has an interceptor that clones requests and retries once through `/api/auth/refresh` on a 401. Types are auto-generated in `src/api/schema.d.ts` — run `pnpm api:generate` after backend changes. API fields must be camelCase; notify the user if any are not.
- **UI:** [shadcn/ui](https://ui.shadcn.com) in `src/components/ui/`, built on `@base-ui/react`. Use `cn()` from `src/lib/utils.ts` for conditional class merging.
- **Styling:** Tailwind CSS v4.
- **Forms/validation:** `@tanstack/react-form` with `zod`. Toasts via `sonner`. Dates/times via `temporal-polyfill`.

### Type Safety Flow

Backend Go types → Huma generates OpenAPI spec → `pnpm api:generate` → `src/api/schema.d.ts` → `openapi-fetch` client gives end-to-end type safety across the API boundary.

## Conventions

- Commits are formatted/linted on staging via `husky` + `lint-staged` (prettier + eslint on `*.{ts,tsx}`, prettier on `*.{json,css,md}`). Keep code passing `pnpm lint`.
- **Comments are rare.** Default to none. Self-explanatory code — clear names, small functions, obvious structure — always beats a comment. If code needs a comment to be understood, rewrite the code first.
- A comment may only explain **why**: a non-obvious tradeoff, a workaround, a constraint that isn't visible in the code. Never write a comment that describes **what** the code does or **how** it does it — that's what the code is for.
