# syntax=docker/dockerfile:1

FROM debian

WORKDIR /app

COPY build/cosmos .

VOLUME /config

EXPOSE 443 80

CMD ["./cosmos"]
