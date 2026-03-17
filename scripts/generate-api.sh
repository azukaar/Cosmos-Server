#!/bin/bash
set -e

cd "$(dirname "$0")/.."

VERSION=$(jq -r .version package.json)

echo "Generating API spec and Go SDK for v${VERSION}..."

# 1. Stamp version into swagger annotation
sed -i "s/@version .*/@version $VERSION/" src/index.go

# 2. Generate Swagger 2.0 spec (using go run — no install needed)
go run github.com/swaggo/swag/cmd/swag@latest init -d src -g index.go -o api-docs --outputTypes json,yaml --parseInternal --parseDependency

# 3. Convert to OpenAPI 3.0
npx swagger2openapi api-docs/swagger.json -o api-docs/openapi.json

# 4. Patch OpenAPI spec: replace Go stdlib types with simple schemas
#    (os.FileMode, time.Duration, nat.PortSet produce uncompilable generated code)
jq '
  .components.schemas["os.FileMode"] = {"type": "integer", "format": "uint32"} |
  .components.schemas["time.Duration"] = {"type": "integer", "format": "int64"} |
  .components.schemas["nat.PortSet"] = {"type": "object", "additionalProperties": true}
' api-docs/openapi.json > api-docs/openapi.patched.json
mv api-docs/openapi.patched.json api-docs/openapi.json

# 5. Generate Go SDK client (using go run — no install needed)
cd go-sdk
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest --config oapi-codegen.yaml ../api-docs/openapi.json

# 6. Stamp Go SDK version
sed -i "s/const Version = \".*\"/const Version = \"$VERSION\"/" version.go

# 7. Resolve dependencies and verify it compiles
go mod tidy
go build ./...

echo "API spec and Go SDK regenerated for v${VERSION}"
