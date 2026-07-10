package bitculator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// APIError is a non-2xx response, carrying the API's
// { error: { code, message, details } } envelope. Inspect it with errors.As:
//
//	var apiErr *bitculator.APIError
//	if errors.As(err, &apiErr) { switch apiErr.Status { ... } }
type APIError struct {
	Status     int
	Code       string
	Message    string
	Details    json.RawMessage
	RequestID  string
	RetryAfter int // seconds; 0 if not provided (429 only)
}

func (e *APIError) Error() string {
	return fmt.Sprintf("bitculator: HTTP %d: %s", e.Status, e.Message)
}

func (e *APIError) IsAuth() bool        { return e.Status == http.StatusUnauthorized }
func (e *APIError) IsPermission() bool  { return e.Status == http.StatusForbidden }
func (e *APIError) IsNotFound() bool    { return e.Status == http.StatusNotFound }
func (e *APIError) IsValidation() bool  { return e.Status == http.StatusUnprocessableEntity }
func (e *APIError) IsRateLimited() bool { return e.Status == http.StatusTooManyRequests }
func (e *APIError) IsServer() bool      { return e.Status >= 500 }

// TimeoutError is returned when a request exceeds its deadline (after retries).
type TimeoutError struct {
	Path string
	Err  error
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("bitculator: request to %s timed out", e.Path)
}
func (e *TimeoutError) Unwrap() error { return e.Err }

// DecodeError is returned when a 2xx response body cannot be decoded into the
// expected shape (malformed JSON, or rows that are not an array).
type DecodeError struct {
	Path string
	Err  error
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("bitculator: decoding response from %s: %v", e.Path, e.Err)
}
func (e *DecodeError) Unwrap() error { return e.Err }

// ConnectionError is returned for network-level failures.
type ConnectionError struct {
	Path string
	Err  error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("bitculator: network error requesting %s: %v", e.Path, e.Err)
}
func (e *ConnectionError) Unwrap() error { return e.Err }

func newAPIError(status int, body []byte, h http.Header) *APIError {
	e := &APIError{Status: status, Message: fmt.Sprintf("HTTP %d", status)}

	if len(body) > 0 {
		var env struct {
			Error *struct {
				Code    string          `json:"code"`
				Message string          `json:"message"`
				Details json.RawMessage `json:"details"`
			} `json:"error"`
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &env) == nil {
			if env.Error != nil {
				e.Code = env.Error.Code
				if env.Error.Message != "" {
					e.Message = env.Error.Message
				}
				e.Details = env.Error.Details
			} else if env.Message != "" {
				e.Message = env.Message
			}
		}
	}

	e.RequestID = h.Get("X-Request-Id")
	if status == http.StatusTooManyRequests {
		if secs, err := strconv.Atoi(h.Get("Retry-After")); err == nil {
			e.RetryAfter = secs
		}
	}
	return e
}
