FROM node:lts-alpine AS build

ENV NPM_CONFIG_UPDATE_NOTIFIER=false
ENV NPM_CONFIG_FUND=false

WORKDIR /app

COPY package.json ./
COPY pnpm-lock.yaml ./

RUN pnpm ci

RUN pnpm build

FROM caddy

WORKDIR /app

COPY Caddyfile ./

RUN caddy fmt Caddyfile --overwrite
COPY --from=build /app/dist ./dist

CMD ["caddy", "run", "--config", "Caddyfile", "--adapter", "caddyfile"]