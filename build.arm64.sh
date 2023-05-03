#!/bin/bash

rm -rf build
env GOARCH=arm64 go build -o build/cosmos src/*.go
if [ $? -ne 0 ]; then
    exit 1
fi
cp -r static build/
cp -r GeoLite2-Country.mmdb build/
cp -r Logo.png build/
mkdir build/images
cp client/src/assets/images/icons/cosmos_gray.png build/cosmos_gray.png
cp client/src/assets/images/icons/cosmos_gray.png cosmos_gray.png
echo '{' > build/meta.json
cat package.json | grep -E '"version"' >> build/meta.json
echo '  "buildDate": "'`date`'",' >> build/meta.json
echo '  "built from": "'`hostname`'"' >> build/meta.json
echo '}' >> build/meta.json