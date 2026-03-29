# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Scope

**Only contribute to the frontend (`web/`).** The backend (`api/`) is Go code that the user is writing themselves as a learning exercise — do not suggest, write, or modify any Go code.

## Frontend Commands (run from `web/`)

```bash
pnpm dev          # Dev server on port 3000
pnpm build        # Production build
pnpm test         # Run Vitest tests
pnpm lint         # ESLint
pnpm check        # Format + lint fix
pnpm api:generate # Regenerate TypeScript types from backend OpenAPI spec
```

## Frontend Architecture (`web/`)

- **Framework:** [TanStack Start](https://tanstack.com/start) (React meta-framework) with Vite.
- **Routing:** TanStack Router with file-based routing in `src/routes/`. The route tree (`routeTree.gen.ts`) is auto-generated — do not edit it manually.
- **Data fetching:** TanStack Query (v5). `QueryClient` is injected into the router context and available in all routes via `route.useRouteContext()`.
- **API client:** `openapi-fetch` with types auto-generated from the backend's OpenAPI spec. Client is in `src/api/client.ts`, types in `src/api/schema.d.ts`. Run `pnpm api:generate` after backend changes to refresh types.
- **UI components:** Shadcn/ui in `src/components/ui/`. Use `cn()` from `src/lib/utils.ts` for conditional class merging.
- **Styling:** Tailwind CSS v4.
- **Data loading pattern:** Route loaders use `queryClient.ensureQueryData()` to prefetch before render, then `useSuspenseQuery()` in the component.

### Type Safety Flow
Backend Go types → Huma generates OpenAPI spec → `pnpm api:generate` → `src/api/schema.d.ts` → `openapi-fetch` client provides end-to-end type safety.

### Data Model
```sql
newsletters (
  id UUID, name TEXT, frequency ENUM('monthly','weekly','daily'),
  send_day INT, send_hour INT, send_minute INT,
  last_sent_at TIMESTAMPTZ, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
)
```
