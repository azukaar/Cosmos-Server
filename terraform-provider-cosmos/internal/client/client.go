package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
)

// CosmosClient wraps the generated SDK client with convenience helpers.
type CosmosClient struct {
	Raw     *cosmossdk.Client
	BaseURL string
}

// apiResponse is the standard Cosmos API envelope.
type apiResponse struct {
	Status  string          `json:"status"`
	Data    json.RawMessage `json:"data,omitempty"`
	Message string          `json:"message,omitempty"`
	Code    string          `json:"code,omitempty"`
}

// retryTransport wraps an http.RoundTripper and retries on connection errors.
type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
	delay      time.Duration
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 0; i <= t.maxRetries; i++ {
		resp, err = t.base.RoundTrip(req)
		if err == nil {
			return resp, nil
		}
		// Only retry on connection errors (server restarting, etc.)
		if !isConnectionError(err) {
			return resp, err
		}
		if i < t.maxRetries {
			time.Sleep(t.delay)
		}
	}
	return resp, err
}

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "connection refused") ||
		strings.Contains(s, "connection reset") ||
		strings.Contains(s, "EOF") ||
		strings.Contains(s, "no such host")
}

// NewCosmosClient creates a new SDK client configured with auth and TLS settings.
func NewCosmosClient(baseURL, token string, insecure bool) (*CosmosClient, error) {
	var base http.RoundTripper = http.DefaultTransport
	if insecure {
		base = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	httpClient := &http.Client{
		Transport: &retryTransport{
			base:       base,
			maxRetries: 3,
			delay:      1 * time.Second,
		},
	}

	opts := []cosmossdk.ClientOption{
		cosmossdk.WithHTTPClient(httpClient),
	}
	if token != "" {
		addToken := func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+token)
			return nil
		}
		opts = append(opts, cosmossdk.WithRequestEditorFn(addToken))
	}

	sdkClient, err := cosmossdk.NewClient(
		baseURL+"/cosmos",
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("creating cosmos SDK client: %w", err)
	}

	return &CosmosClient{
		Raw:     sdkClient,
		BaseURL: baseURL,
	}, nil
}

// NotFoundError is returned when the API returns HTTP 404.
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// IsNotFound returns true if the error is a 404 response.
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// ParseResponse reads an http.Response body, checks the status envelope, and
// unmarshals the data field into T.
func ParseResponse[T any](resp *http.Response) (*T, error) {
	if resp == nil {
		return nil, fmt.Errorf("nil response")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode == 404 {
		return nil, &NotFoundError{Message: string(body)}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var envelope apiResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("parsing response envelope: %w (body: %s)", err, string(body))
	}

	if envelope.Status != "OK" && envelope.Status != "ok" {
		msg := envelope.Message
		if msg == "" {
			msg = string(body)
		}
		return nil, fmt.Errorf("API error (status=%s): %s", envelope.Status, msg)
	}

	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil, nil
	}

	var result T
	if err := json.Unmarshal(envelope.Data, &result); err != nil {
		return nil, fmt.Errorf("parsing response data: %w (raw: %s)", err, string(envelope.Data))
	}

	return &result, nil
}

// CheckResponse reads an http.Response and verifies the status envelope without
// trying to parse data.
func CheckResponse(resp *http.Response) error {
	if resp == nil {
		return fmt.Errorf("nil response")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode == 404 {
		return &NotFoundError{Message: string(body)}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var envelope apiResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("parsing response envelope: %w (body: %s)", err, string(body))
	}

	if envelope.Status != "OK" && envelope.Status != "ok" {
		msg := envelope.Message
		if msg == "" {
			msg = string(body)
		}
		return fmt.Errorf("API error (status=%s): %s", envelope.Status, msg)
	}

	return nil
}

// ParseRawResponse reads an http.Response and returns the raw data field from
// the response envelope.
func ParseRawResponse(resp *http.Response) (json.RawMessage, error) {
	if resp == nil {
		return nil, fmt.Errorf("nil response")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode == 404 {
		return nil, &NotFoundError{Message: string(body)}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var envelope apiResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("parsing response envelope: %w (body: %s)", err, string(body))
	}

	if envelope.Status != "OK" && envelope.Status != "ok" {
		msg := envelope.Message
		if msg == "" {
			msg = string(body)
		}
		return nil, fmt.Errorf("API error (status=%s): %s", envelope.Status, msg)
	}

	return envelope.Data, nil
}

// CheckStreamingResponse handles API endpoints that stream progress text
// (e.g. docker pull logs, service creation) instead of returning a standard
// JSON envelope. It reads the full body and checks for success/failure markers.
func CheckStreamingResponse(resp *http.Response) error {
	if resp == nil {
		return fmt.Errorf("nil response")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	bodyStr := string(body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, bodyStr)
	}

	// Check for failure indicators in the streamed output.
	if strings.Contains(bodyStr, "[OPERATION FAILED]") || strings.Contains(bodyStr, "[FAIL]") {
		return fmt.Errorf("operation failed: %s", bodyStr)
	}

	return nil
}

// Helper functions for converting between Terraform and SDK types.

func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func StringPtrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func BoolPtr(b bool) *bool {
	return &b
}

func BoolPtrVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func IntPtr(i int) *int {
	return &i
}

func IntPtrVal(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func Int64Ptr(i int64) *int64 {
	return &i
}

func Int64PtrVal(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func Float32Ptr(f float32) *float32 {
	return &f
}

func Float32PtrVal(f *float32) float32 {
	if f == nil {
		return 0
	}
	return *f
}
