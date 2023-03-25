# syntax=docker/dockerfile:1

FROM debian

EXPOSE 443 80

VOLUME /config

RUN apt-get update && apt-get install -y ca-certificates openssl

WORKDIR /app

COPY build/cosmos .
COPY static ./static

CMD ["./cosmos"]
