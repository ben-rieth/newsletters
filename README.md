# Newsletters

A self-hosted newsletter service that aggregates RSS/Atom feeds into scheduled email digests. Create newsletters, add feeds, filter content, and deliver curated digests on your own schedule.

## Features

- **Newsletter management** — create multiple newsletters with independent schedules (daily, weekly, monthly) and timezone support
- **RSS/Atom aggregation** — add feeds by URL; content is fetched and stored automatically
- **Feed filtering** — include or exclude items by title or URL patterns
- **Preview** — see filtered feed output before a newsletter goes out
- **Scheduled delivery** — a built-in scheduler sends newsletters at configured times via [Resend](https://resend.com)
- **Auth** — JWT-based auth with email verification and token refresh/revocation
- **Unsubscribe** — public unsubscribe links included in every sent newsletter

## Stack

| Layer | Technology |
|---|---|
| Backend | Go, [Huma v2](https://huma.rocks) (REST + OpenAPI), PostgreSQL |
| Frontend | React, [TanStack Start](https://tanstack.com/start), TanStack Router, TanStack Query v5 |
| API client | [openapi-fetch](https://openapi-ts.dev/openapi-fetch/) with auto-generated types |
| UI | [shadcn/ui](https://ui.shadcn.com), Tailwind CSS v4 |
| Email | [Resend](https://resend.com) |

## Project Structure

```
├── api/          # Go backend (port 8080)
│   ├── cmd/      # Entry point
│   ├── db/       # Schema, migrations, sqlc queries
│   └── internal/ # Handlers, services (auth, feeds, email, newsletters)
└── web/          # React frontend (port 3000)
    └── src/
        ├── api/      # openapi-fetch client + generated types
        ├── features/ # Feature modules (auth, newsletters, unsubscribe)
        ├── routes/   # File-based TanStack Router pages
        └── components/ui/  # shadcn/ui components
```

## Prerequisites

- Go 1.25+
- Node.js + [pnpm](https://pnpm.io)
- PostgreSQL
- [Air](https://github.com/air-verse/air) (Go live reloader) — `go install github.com/air-verse/air@latest`
- A [Resend](https://resend.com) API key

## Getting Started

**1. Clone and set up the database**

```bash
git clone <repo-url>
cd newsletters
# Create a PostgreSQL database and run api/db/schema.sql
psql -d your_db < api/db/schema.sql
```

**2. Configure the backend**

```bash
cp api/.env.example api/.env  # or create api/.env manually
```

```env
# api/.env
DATABASE_URL=postgres://user:password@localhost:5432/newsletters
JWT_SECRET=your-secret-here
RESEND_API_KEY=re_...
EMAIL_FROM=newsletters@yourdomain.com
```

**3. Configure the frontend**

```bash
cp web/.env.example web/.env  # or create web/.env manually
```

```env
# web/.env
VITE_API_URL=http://localhost:8080
```

**4. Install frontend dependencies**

```bash
cd web && pnpm install
```

**5. Start development servers**

```bash
# From the repo root — runs both API and web in parallel
make dev
```

- API: http://localhost:8080
- Frontend: http://localhost:3000
- OpenAPI spec: http://localhost:8080/openapi.json

## Frontend Commands

Run from `web/`:

```bash
pnpm dev           # Dev server
pnpm build         # Production build
pnpm test          # Vitest
pnpm lint          # ESLint
pnpm check         # Format + lint fix
pnpm api:generate  # Regenerate TypeScript types from backend OpenAPI spec
```

## Regenerating API Types

After changing backend types, refresh the frontend's TypeScript types:

```bash
cd web && pnpm api:generate
```

This fetches the OpenAPI spec from the running backend and regenerates `src/api/schema.d.ts`.
