# Cosmos Server Go SDK

Auto-generated Go client for the [Cosmos Server](https://github.com/azukaar/cosmos-server) API.

## Installation

```bash
go get github.com/azukaar/cosmos-server/go-sdk
```

## Usage

```go
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
)

func main() {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	addToken := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}

	client, err := cosmossdk.NewClient(
		"https://cosmos.example.com/cosmos",
		cosmossdk.WithHTTPClient(httpClient),
		cosmossdk.WithRequestEditorFn(addToken),
	)
	if err != nil {
		panic(err)
	}

	// List routes
	resp, err := client.GetApiRoutes(context.Background())
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)
}
```

## Generation

The SDK is auto-generated from the OpenAPI spec using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen). To regenerate:

```bash
bash scripts/generate-api.sh
```

This runs on every commit via a pre-commit hook.
