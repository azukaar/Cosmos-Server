FROM --platform=$BUILDPLATFORM node:21 AS node-builder
WORKDIR /usr/src/app

RUN mkdir -p build && \
    printf '{"version": "%s", "buildDate": "%s", "built from": "%s"}' \
        $(cat package.json | grep "version" | cut -d'"' -f 4) \
        $(date "+%F-%H-%M-%S") \
        $(hostname) > build/meta.json

COPY package.json package-lock.json .
ENV npm_config_cache=/root/.cache/npm
RUN --mount=type=cache,target="$npm_config_cache" npm install

COPY . .
RUN --mount=type=cache,target=./node_modules/.cache/webpack npm run webpack:build

FROM --platform=$BUILDPLATFORM golang:1.21.5-alpine AS go-builder
WORKDIR /usr/src/app

ENV GOMODCACHE=/root/.cache/gomod-cache
ENV GOCACHE=/root/.cache/go-build

COPY go.mod go.sum .
RUN --mount=type=cache,target="$GOMODCACHE" go mod download

COPY . .

ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target="$GOCACHE" --mount=type=cache,target="$GOMODCACHE" GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o build/cosmos src/*.go

FROM alpine:3.19.0

RUN apk --no-cache add ca-certificates openssl

WORKDIR /app

COPY --from=node-builder /usr/src/app/static static
COPY --from=node-builder /usr/src/app/build .
COPY --from=node-builder /usr/src/app/client/src/assets/images/icons/cosmos_gray.png .
COPY --from=go-builder /usr/src/app/build .

VOLUME /config

EXPOSE 443 80 4242/udp
CMD ./cosmos
