FROM node:lts-alpine AS base

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

ENV CI=true
ENV HUSKY=0

FROM base AS prod
WORKDIR /app
COPY pnpm-lock.yaml ./
RUN --mount=type=cache,id=pnpm,target=/pnpm/store \
    pnpm fetch
COPY package.json ./
RUN --mount=type=cache,id=pnpm,target=/pnpm/store \
    pnpm install --offline

COPY . .
RUN pnpm build

FROM caddy:2-alpine

WORKDIR /app

COPY Caddyfile ./

RUN caddy fmt Caddyfile --overwrite
COPY --from=prod /app/dist /app/dist

CMD ["caddy", "run", "--config", "Caddyfile", "--adapter", "caddyfile"]
