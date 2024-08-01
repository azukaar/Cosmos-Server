# syntax=docker/dockerfile:1

FROM debian:12

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

RUN apt-get update \
    && apt-get install -y ca-certificates openssl fdisk mergerfs snapraid \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the respective binary based on the BINARY_NAME
COPY build/cosmos build/cosmos-arm64 ./

# Copy other resources
COPY build/* ./
COPY static ./static

# Run the respective binary based on the BINARY_NAME
CMD ["sh", "-c", "./$(cat /binary_name)"]