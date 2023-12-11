FROM node:21 AS node-builder
WORKDIR /usr/src/app

COPY package.json package-lock.json .
RUN npm install

COPY . .
RUN mkdir -p build && \
    printf '{"version": "%s", "buildDate": "%s", "built from": "%s"}' \
        $(cat package.json | grep "version" | cut -d'"' -f 4) \
        $(date "+%F-%H-%M-%S") \
        $(hostname) > build/meta.json
RUN npm run webpack:build

FROM golang:1.21.5 AS go-builder
WORKDIR /usr/src/app

COPY go.mod go.sum .
RUN go mod download && go mod verify

COPY . .
RUN go build -o build/cosmos src/*.go

FROM alpine:3.19.0

RUN apk --no-cache add libc6-compat ca-certificates openssl

WORKDIR /app

COPY --from=node-builder /usr/src/app/static static
COPY --from=node-builder /usr/src/app/build .
COPY --from=node-builder /usr/src/app/client/src/assets/images/icons/cosmos_gray.png .
COPY --from=go-builder /usr/src/app/build .

VOLUME /config

EXPOSE 443 80
CMD ./cosmos
