default: build

build:
	go build -o terraform-provider-cosmos

install: build
	mkdir -p ~/.terraform.d/plugins/cosmos-cloud.io/azukaar/cosmos/0.1.0/linux_amd64
	cp terraform-provider-cosmos ~/.terraform.d/plugins/cosmos-cloud.io/azukaar/cosmos/0.1.0/linux_amd64/

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./... -v -timeout 120m

fmt:
	go fmt ./...

.PHONY: build install test testacc fmt
