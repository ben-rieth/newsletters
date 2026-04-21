FROM node:lts-alpine AS base

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

ENV CI=true
ENV HUSKY=0

FROM base AS prod
WORKDIR /app
COPY pnpm-lock.yaml ./
RUN pnpm fetch
COPY package.json ./
RUN pnpm install --offline

COPY . .
ARG VITE_API_URL
RUN pnpm build

FROM caddy:2-alpine

WORKDIR /app

COPY Caddyfile ./

RUN caddy fmt Caddyfile --overwrite
COPY --from=prod /app/dist /app/dist

CMD ["caddy", "run", "--config", "Caddyfile", "--adapter", "caddyfile"]
