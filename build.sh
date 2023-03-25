rm -rf build
go build -o build/cosmos src/*.go
if [ $? -ne 0 ]; then
    exit 1
fi
cp -r static build/