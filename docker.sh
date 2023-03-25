VERSION=$(npm pkg get version | tr -d \")

if [ -n "$ARCHI" ]; then
  VERSION="$ARCHI-$VERSION"
fi

echo "Pushing azukaar/cosmos-server:$VERSION to docker hub"

sh build.sh

docker build \
  -t azukaar/cosmos-server:$VERSION \
  -t azukaar/cosmos-server:latest \
  .

sh build arm64.sh

docker build \
  -t azukaar/cosmos-server:$VERSION-arm64 \
  -t azukaar/cosmos-server:latest-arm64 \
  -f Dockerfile.arm64 \
  --platform linux/arm64 \
  .

docker push azukaar/cosmos-server:$VERSION
docker push azukaar/cosmos-server:latest
docker push azukaar/cosmos-server:$VERSION-arm64
docker push azukaar/cosmos-server:latest-arm64