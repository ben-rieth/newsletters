FROM node:lts-alpine AS base

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable pnpm && corepack install -g pnpm@latest-11

ENV CI=true
ENV HUSKY=0

FROM base AS prod
WORKDIR /app
COPY pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm fetch
COPY package.json ./
RUN pnpm install --offline

ARG VITE_API_URL
ARG WEB_URL
ENV VITE_API_URL=$VITE_API_URL
ENV WEB_URL=$WEB_URL

COPY . .
RUN pnpm build

FROM caddy:2-alpine

WORKDIR /app

COPY Caddyfile ./

RUN caddy fmt Caddyfile --overwrite
COPY --from=prod /app/dist /app/dist

CMD ["caddy", "run", "--config", "Caddyfile", "--adapter", "caddyfile"]
