rm -rf build
go build -o build/cosmos src/*.go
if [ $? -ne 0 ]; then
    exit 1
fi
cp -r static build/
echo '{' > build/meta.json
cat package.json | grep -E '"version"' >> build/meta.json
echo '  "buildDate": "'`date`'",' >> build/meta.json
echo '  "built from": "'`hostname`'"' >> build/meta.json
echo '}' >> build/meta.json