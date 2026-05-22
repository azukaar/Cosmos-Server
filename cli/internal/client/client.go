// Package client provides a configured Cosmos Server API client.
package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
	"github.com/azukaar/cosmos-server/cli/internal/config"
)

// APIResponse is the standard Cosmos Server response envelope.
type APIResponse struct {
	Status  string          `json:"status"`
	Data    json.RawMessage `json:"data,omitempty"`
	Message string          `json:"message,omitempty"`
	Code    string          `json:"code,omitempty"`
	Total   int             `json:"total,omitempty"`
}

// New creates a new Cosmos Server API client from a resolved config.
func New(r *config.Resolved) (*cosmossdk.Client, error) {
	token := r.Token
	host := r.Host

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	addHeaders := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Host", host)
		return nil
	}

	serverURL := r.URL + "/cosmos"
	return cosmossdk.NewClient(serverURL,
		cosmossdk.WithHTTPClient(httpClient),
		cosmossdk.WithRequestEditorFn(addHeaders),
	)
}

// Parse reads and unmarshals a Cosmos API response.
func Parse(resp *http.Response) (*APIResponse, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var r APIResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("parsing response (status %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode >= 400 {
		msg := r.Message
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
	}

	return &r, nil
}
