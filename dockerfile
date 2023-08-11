# syntax=docker/dockerfile:1

FROM debian:11

ARG TARGETPLATFORM
ARG BINARY_NAME=cosmos

# Set BINARY_NAME based on the TARGETPLATFORM
RUN case "$TARGETPLATFORM" in \
    "linux/arm64") BINARY_NAME="cosmos-arm64" ;; \
    *) BINARY_NAME="cosmos" ;; \
    esac && echo $BINARY_NAME > /binary_name

# This is just to log the platforms (optional)
RUN echo "I am building for $TARGETPLATFORM" > /log

EXPOSE 443 80

VOLUME /config

RUN apt-get update && apt-get install -y ca-certificates openssl

WORKDIR /app

# Copy the respective binary based on the BINARY_NAME
COPY build/$BINARY_NAME ./

# Copy other resources
COPY build/cosmos_gray.png build/Logo.png build/GeoLite2-Country.mmdb build/meta.json ./
COPY static ./static

# Run the respective binary based on the BINARY_NAME
CMD ["sh", "-c", "./$(cat /binary_name)"]