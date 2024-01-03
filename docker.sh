#!/bin/bash

VERSION=$(cat package.json | grep "version" | cut -d'"' -f 4)
LATEST="latest"

# if branch is unstable in git for circle ci
if [ -n "$CIRCLE_BRANCH" ]; then
  if [ "$CIRCLE_BRANCH" != "master" ]; then
    LATEST="$LATEST-$CIRCLE_BRANCH"
  fi
fi

echo "Pushing azukaar/cosmos-server:$VERSION and azukaar/cosmos-server:$LATEST"

# Multi-architecture build
docker buildx build \
  --platform linux/amd64,linux/arm64,linux/i386,linux/ppc64le,linux/s390x \
  --tag azukaar/cosmos-server:$VERSION \
  --tag azukaar/cosmos-server:$LATEST \
  --push \
  .
