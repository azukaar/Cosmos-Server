# syntax=docker/dockerfile:1

FROM debian

WORKDIR /app

COPY build/cosmos .
COPY static ./static

VOLUME /config

EXPOSE 443 80

CMD ["./cosmos"]
