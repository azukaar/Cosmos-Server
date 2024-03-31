ARG NODEJS_IMAGE=docker.io/node:21.7.1-slim
ARG GO_IMAGE=docker.io/golang:1.22.1
ARG BASE_IMAGE=docker.io/bitnami/minideb:latest
ARG CACHE_DIR=/root/.cache
ARG APP_BUILD_DIR=/usr/src/app
ARG APP_TARGET_DIR=/usr

FROM --platform=$BUILDPLATFORM $NODEJS_IMAGE AS node-builder
ARG APP_BUILD_DIR NODE_ENV CACHE_DIR
WORKDIR $APP_BUILD_DIR

COPY package.json package-lock.json .
ENV npm_config_cache=$CACHE_DIR/npm
RUN --mount=type=cache,target=$npm_config_cache \
  npm ci || (npm cache clean --force && npm ci)

COPY vite.config.js .
COPY client client
RUN --mount=type=cache,target=./node_modules/.vite \
  npm run build:client

FROM --platform=$BUILDPLATFORM $GO_IMAGE AS app-builder
ARG APP_BUILD_DIR TARGETOS TARGETARCH CACHE_DIR
WORKDIR $APP_BUILD_DIR

ENV GOMODCACHE=$CACHE_DIR/go-mod

COPY go.mod go.sum .
RUN --mount=type=cache,target=$GOMODCACHE \
  go mod download && go mod verify

COPY src src
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH GOCACHE=$CACHE_DIR/go-build
RUN --mount=type=cache,target=$GOCACHE --mount=type=cache,target=$GOMODCACHE \
  go build -o build/cosmos src/*.go

FROM $BASE_IMAGE
ARG APP_TARGET_DIR APP_BUILD_DIR
WORKDIR $APP_TARGET_DIR

RUN --mount=type=cache,target=/var/cache/apt \
  install_packages ca-certificates openssl fdisk mergerfs snapraid

COPY --from=node-builder $APP_BUILD_DIR/static static
COPY --from=app-builder $APP_BUILD_DIR/build .

VOLUME /config

ENV APP_BIN=$APP_TARGET_DIR/cosmos

EXPOSE 443 80 4242/udp
ENTRYPOINT $APP_BIN
CMD []
