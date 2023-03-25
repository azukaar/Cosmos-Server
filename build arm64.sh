rm -rf build
env GOARCH=arm64 go build -o build/cosmos src/*.go
if [ $? -ne 0 ]; then
    exit 1
fi
cp -r static build/