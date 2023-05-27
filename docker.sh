#!/bin/bash

VERSION=$(grep -o '\"version\": \"[^\"]*\"' package.json | sed 's/[^0-9.]//g')
LATEST="latest"

# if branch is unstable in git for circle ci
if [ -n "$CIRCLE_BRANCH" ]; then
  if [ "$CIRCLE_BRANCH" != "master" ]; then
    LATEST="$LATEST-$CIRCLE_BRANCH"
  fi
fi

echo "Pushing azukaar/cosmos-server:$VERSION and azukaar/cosmos-server:$LATEST"

docker build \
  -t azukaar/cosmos-server:$VERSION \
  -t azukaar/cosmos-server:$LATEST \
  .

docker push azukaar/cosmos-server:$VERSION
docker push azukaar/cosmos-server:$LATEST