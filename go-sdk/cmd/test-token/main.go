// API Token Integration Test Script using the Go SDK
// Usage: go run ./cmd/test-token <base_url> <admin_token> <readonly_token>
// Example: cd go-sdk && go run ./cmd/test-token https://localhost cosmos_abc123 cosmos_xyz789
//
// Accepts self-signed certs.
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	cosmossdk "github.com/azukaar/cosmos-server/go-sdk"
)

var (
	passed int
	failed int
)

// apiResponse is the standard Cosmos response shape
type apiResponse struct {
	Status  string          `json:"status"`
	Data    json.RawMessage `json:"data,omitempty"`
	Message string          `json:"message,omitempty"`
	Code    string          `json:"code,omitempty"`
}

func test(name string, fn func() error) {
	if err := fn(); err != nil {
		failed++
		fmt.Printf("[FAIL] %s — %s\n", name, err)
	} else {
		passed++
		fmt.Printf("[PASS] %s\n", name)
	}
}

func expectStatus(resp *http.Response, expected int, label string) error {
	if resp.StatusCode != expected {
		return fmt.Errorf("%s: expected %d, got %d", label, expected, resp.StatusCode)
	}
	return nil
}

func parseBody(resp *http.Response) (*apiResponse, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var r apiResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(body))
	}
	return &r, nil
}

func expectOK(resp *http.Response) error {
	r, err := parseBody(resp)
	if err != nil {
		return err
	}
	if r.Status != "OK" {
		return fmt.Errorf("expected status OK, got %s (message: %s)", r.Status, r.Message)
	}
	return nil
}

func expectForbiddenOrError(resp *http.Response) error {
	if resp.StatusCode == 403 {
		resp.Body.Close()
		return nil
	}
	r, err := parseBody(resp)
	if err != nil {
		return err
	}
	if r.Status == "OK" {
		return fmt.Errorf("expected error/forbidden, but got OK")
	}
	return nil
}

func newClient(baseUrl, token string) *cosmossdk.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	addToken := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}

	c, err := cosmossdk.NewClient(baseUrl+"/cosmos", cosmossdk.WithHTTPClient(httpClient), cosmossdk.WithRequestEditorFn(addToken))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}
	return c
}

func rawRequest(baseUrl, path, method, token string, body interface{}) (*http.Response, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = strings.NewReader(string(b))
	}

	req, err := http.NewRequest(method, baseUrl+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return httpClient.Do(req)
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: go run test-token.go <base_url> <admin_token> <readonly_token>\n")
		os.Exit(2)
	}

	baseUrl := os.Args[1]
	adminToken := os.Args[2]
	readonlyToken := os.Args[3]

	admin := newClient(baseUrl, adminToken)
	readonly := newClient(baseUrl, readonlyToken)
	ctx := context.Background()

	fmt.Printf("Testing API tokens via Go SDK against %s\n", baseUrl)
	fmt.Printf("Admin token: %s...\n", adminToken[:12])
	fmt.Printf("Readonly token: %s...\n", readonlyToken[:12])

	// --- Group 1: SDK Basic Calls ---
	fmt.Println("\n— Group 1: SDK Basic Calls —")

	test("admin.GetApiStatus()", func() error {
		resp, err := admin.GetApiStatus(ctx)
		if err != nil {
			return err
		}
		return expectOK(resp)
	})

	test("readonly.GetApiStatus()", func() error {
		resp, err := readonly.GetApiStatus(ctx)
		if err != nil {
			return err
		}
		return expectOK(resp)
	})

	// --- Group 2: SDK Read Endpoints ---
	fmt.Println("\n— Group 2: SDK Read Endpoints —")

	test("admin.GetApiConfig()", func() error {
		resp, err := admin.GetApiConfig(ctx)
		if err != nil {
			return err
		}
		return expectOK(resp)
	})

	test("readonly.GetApiConfig()", func() error {
		resp, err := readonly.GetApiConfig(ctx)
		if err != nil {
			return err
		}
		return expectOK(resp)
	})

	test("admin.GetApiServapps()", func() error {
		resp, err := admin.GetApiServapps(ctx, nil)
		if err != nil {
			return err
		}
		return expectOK(resp)
	})

	test("readonly.GetApiServapps()", func() error {
		resp, err := readonly.GetApiServapps(ctx, nil)
		if err != nil {
			return err
		}
		return expectOK(resp)
	})

	test("admin.GetApiUsers()", func() error {
		resp, err := admin.GetApiUsers(ctx, nil)
		if err != nil {
			return err
		}
		return expectOK(resp)
	})

	test("readonly.GetApiUsers()", func() error {
		resp, err := readonly.GetApiUsers(ctx, nil)
		if err != nil {
			return err
		}
		return expectOK(resp)
	})

	test("admin.GetApiApiTokens()", func() error {
		resp, err := admin.GetApiApiTokens(ctx)
		if err != nil {
			return err
		}
		return expectOK(resp)
	})

	test("readonly.GetApiApiTokens()", func() error {
		resp, err := readonly.GetApiApiTokens(ctx)
		if err != nil {
			return err
		}
		return expectOK(resp)
	})

	// --- Group 3: SDK Write Endpoints (admin vs read-only) ---
	fmt.Println("\n— Group 3: SDK Write Endpoints (admin vs read-only) —")

	probeCreated := false

	test("admin.PostApiApiTokens() - create probe", func() error {
		resp, err := admin.PostApiApiTokens(ctx, cosmossdk.PostApiApiTokensJSONRequestBody{
			Name: "__test_probe__",
		})
		if err != nil {
			return err
		}
		if err := expectOK(resp); err != nil {
			return err
		}
		probeCreated = true
		return nil
	})

	test("admin.DeleteApiApiTokens() - cleanup probe", func() error {
		resp, err := admin.DeleteApiApiTokens(ctx, cosmossdk.DeleteApiApiTokensJSONRequestBody{
			Name: "__test_probe__",
		})
		if err != nil {
			return err
		}
		if err := expectOK(resp); err != nil {
			return err
		}
		probeCreated = false
		return nil
	})

	if probeCreated {
		// Safety cleanup
		admin.DeleteApiApiTokens(ctx, cosmossdk.DeleteApiApiTokensJSONRequestBody{
			Name: "__test_probe__",
		})
	}

	test("readonly.PostApiApiTokens() - denied", func() error {
		resp, err := readonly.PostApiApiTokens(ctx, cosmossdk.PostApiApiTokensJSONRequestBody{
			Name: "__test_probe__",
		})
		if err != nil {
			return err
		}
		return expectForbiddenOrError(resp)
	})

	test("readonly.DeleteApiApiTokens() - denied", func() error {
		resp, err := readonly.DeleteApiApiTokens(ctx, cosmossdk.DeleteApiApiTokensJSONRequestBody{
			Name: "__test_probe__",
		})
		if err != nil {
			return err
		}
		return expectForbiddenOrError(resp)
	})

	test("admin.PutApiConfig() - idempotent write-back", func() error {
		getResp, err := admin.GetApiConfig(ctx)
		if err != nil {
			return err
		}
		body, _ := io.ReadAll(getResp.Body)
		getResp.Body.Close()

		var parsed apiResponse
		json.Unmarshal(body, &parsed)

		resp, err := admin.PutApiConfigWithBody(ctx, "application/json", strings.NewReader(string(parsed.Data)))
		if err != nil {
			return err
		}
		return expectOK(resp)
	})

	test("readonly.PutApiConfig() - denied", func() error {
		resp, err := readonly.PutApiConfigWithBody(ctx, "application/json", strings.NewReader("{}"))
		if err != nil {
			return err
		}
		return expectForbiddenOrError(resp)
	})

	// --- Group 4: No-Token Rejection ---
	fmt.Println("\n— Group 4: No-Token Rejection —")

	for _, tc := range []struct {
		path  string
		label string
	}{
		{"/cosmos/api/config", "config"},
		{"/cosmos/api/servapps", "servapps"},
		{"/cosmos/api/users", "users"},
	} {
		label := tc.label
		path := tc.path
		test(fmt.Sprintf("GET %s - no token rejected", label), func() error {
			resp, err := rawRequest(baseUrl, path, "GET", "", nil)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			return expectStatus(resp, 401, "no token")
		})
	}

	// --- Group 5: Invalid Token ---
	fmt.Println("\n— Group 5: Invalid Token —")

	test("status - invalid token rejected", func() error {
		resp, err := rawRequest(baseUrl, "/cosmos/api/status", "GET", "cosmos_invalidtoken", nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return expectStatus(resp, 401, "invalid token")
	})

	// --- Results ---
	fmt.Printf("\nResults: %d/%d passed\n", passed, passed+failed)
	if failed > 0 {
		fmt.Printf("%d test(s) FAILED\n", failed)
		os.Exit(1)
	}
}
