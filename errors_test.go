package bitculator

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

// oneShot builds a client that never retries, so every status maps straight to its error.
func oneShot(t *testing.T, status int, body string, headers map[string]string) *Client {
	t.Helper()
	return newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(w, status, body, headers)
	}), WithMaxRetries(0))
}

func apiErrFrom(t *testing.T, client *Client) *APIError {
	t.Helper()
	_, err := client.Coins.Get(context.Background(), "bitcoin", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %v", err)
	}
	return apiErr
}

func TestStatusPredicates(t *testing.T) {
	cases := []struct {
		status int
		check  func(*APIError) bool
		name   string
	}{
		{401, (*APIError).IsAuth, "IsAuth"},
		{403, (*APIError).IsPermission, "IsPermission"},
		{404, (*APIError).IsNotFound, "IsNotFound"},
		{422, (*APIError).IsValidation, "IsValidation"},
		{429, (*APIError).IsRateLimited, "IsRateLimited"},
		{500, (*APIError).IsServer, "IsServer"},
		{503, (*APIError).IsServer, "IsServer"},
	}
	for _, tc := range cases {
		apiErr := apiErrFrom(t, oneShot(t, tc.status, errorBody("some_code", "nope"), nil))
		if apiErr.Status != tc.status {
			t.Fatalf("status = %d, want %d", apiErr.Status, tc.status)
		}
		if !tc.check(apiErr) {
			t.Fatalf("%s should be true for %d", tc.name, tc.status)
		}
	}
}

func TestCarriesCodeMessageAndDetailsFromTheEnvelope(t *testing.T) {
	body := `{"error":{"code":"validation_failed","message":"Invalid parameters.","details":{"per_page":["must not be greater than 100"]}}}`
	apiErr := apiErrFrom(t, oneShot(t, 422, body, nil))
	if apiErr.Code != "validation_failed" {
		t.Fatalf("code = %q", apiErr.Code)
	}
	if apiErr.Message != "Invalid parameters." {
		t.Fatalf("message = %q", apiErr.Message)
	}
	if len(apiErr.Details) == 0 {
		t.Fatal("details missing")
	}
}

func TestFallsBackToTopLevelMessageThenHTTPStatus(t *testing.T) {
	withMessage := apiErrFrom(t, oneShot(t, 500, `{"message":"Whoops."}`, nil))
	if withMessage.Message != "Whoops." {
		t.Fatalf("message = %q", withMessage.Message)
	}

	rawText := apiErrFrom(t, oneShot(t, 502, `Bad Gateway`, nil))
	if rawText.Message != "HTTP 502" {
		t.Fatalf("message = %q", rawText.Message)
	}
	if rawText.Code != "" {
		t.Fatalf("code = %q", rawText.Code)
	}
}

func TestCapturesTheRequestIDHeader(t *testing.T) {
	apiErr := apiErrFrom(t, oneShot(t, 500, errorBody("server_error", "boom"), map[string]string{"X-Request-Id": "req_123"}))
	if apiErr.RequestID != "req_123" {
		t.Fatalf("request id = %q", apiErr.RequestID)
	}
}

func TestRateLimitCarriesRetryAfter(t *testing.T) {
	apiErr := apiErrFrom(t, oneShot(t, 429, errorBody("rate_limited", "Slow down"), map[string]string{"Retry-After": "30"}))
	if apiErr.RetryAfter != 30 {
		t.Fatalf("retry after = %d", apiErr.RetryAfter)
	}

	absent := apiErrFrom(t, oneShot(t, 429, errorBody("rate_limited", "Slow down"), nil))
	if absent.RetryAfter != 0 {
		t.Fatalf("retry after = %d, want 0 when absent", absent.RetryAfter)
	}
}
