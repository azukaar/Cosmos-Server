# syntax=docker/dockerfile:1

FROM debian:11

EXPOSE 443 80

VOLUME /config

RUN apt-get update && apt-get install -y ca-certificates openssl

WORKDIR /app

COPY build/cosmos .
COPY build/cosmos_gray.png .
COPY build/Logo.png .
COPY build/GeoLite2-Country.mmdb .
COPY build/meta.json .
COPY static ./static

CMD ["./cosmos"]