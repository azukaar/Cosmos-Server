#!/bin/bash

VERSION=$(grep -o '\"version\": \"[^\"]*\"' package.json | sed 's/[^0-9a-z.-]//g'| sed 's/version//g')
LATEST="latest"

# if branch is unstable in git for circle ci
if [ -n "$CIRCLE_BRANCH" ]; then
  if [ "$CIRCLE_BRANCH" != "master" ]; then
    LATEST="$LATEST-$CIRCLE_BRANCH"
  fi
fi

echo "Pushing aseracorp/resiOS:$VERSION and aseracorp/resiOS:$LATEST"

sh build.sh

# Multi-architecture build
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag aseracorp/resiOS:$VERSION \
  --tag aseracorp/resiOS:$LATEST \
  --push \
  .
