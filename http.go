package bitculator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Quota holds the X-Quota-* headers from the most recent response. A field is
// nil when the corresponding header was absent.
type Quota struct {
	Limit     *int
	Remaining *int
	Reset     *int
	Raw       map[string]string
}

// httpClient is the DRY transport core. Every request in the SDK flows through
// do(): URL + query building, auth headers, JSON (de)serialization, the response
// envelope, quota capture, retries with backoff, and error mapping all live here.
type httpClient struct {
	apiKey     string
	baseURL    string
	userAgent  string
	maxRetries int
	hc         *http.Client

	mu    sync.Mutex
	quota *Quota
}

type envelope struct {
	Data  json.RawMessage `json:"data"`
	Meta  *PageMeta       `json:"meta"`
	Links *PageLinks      `json:"links"`
}

func (c *httpClient) get(ctx context.Context, path string, params Params) (json.RawMessage, error) {
	data, _, _, err := c.do(ctx, http.MethodGet, path, params, nil)
	return data, err
}

func (c *httpClient) post(ctx context.Context, path string, body any) (json.RawMessage, error) {
	data, _, _, err := c.do(ctx, http.MethodPost, path, nil, body)
	return data, err
}

func (c *httpClient) delete(ctx context.Context, path string) (json.RawMessage, error) {
	data, _, _, err := c.do(ctx, http.MethodDelete, path, nil, nil)
	return data, err
}

func (c *httpClient) page(ctx context.Context, path string, params Params) (*Page, error) {
	data, meta, links, err := c.do(ctx, http.MethodGet, path, params, nil)
	if err != nil {
		return nil, err
	}
	var rows []json.RawMessage
	if len(data) > 0 {
		if err := json.Unmarshal(data, &rows); err != nil {
			return nil, &DecodeError{Path: path, Err: err}
		}
	}
	p := &Page{Data: rows, c: c, path: path, params: params}
	if meta != nil {
		p.Meta = *meta
	}
	if links != nil {
		p.Links = *links
	}
	return p, nil
}

func (c *httpClient) do(ctx context.Context, method, path string, params Params, body any) (json.RawMessage, *PageMeta, *PageLinks, error) {
	reqURL := c.baseURL + path + encodeParams(params)

	// GET and DELETE are idempotent and safe to auto-retry; POST is not, so a
	// network/timeout failure or a 5xx must never silently re-send it (which
	// could duplicate a resource the server already processed). A 429 is always
	// retryable because the server rejected the request before processing it.
	isIdempotent := method == http.MethodGet || method == http.MethodDelete

	var payload []byte
	if body != nil {
		var err error
		if payload, err = json.Marshal(body); err != nil {
			return nil, nil, nil, fmt.Errorf("bitculator: encoding request body: %w", err)
		}
	}

	for attempt := 0; ; attempt++ {
		var reader io.Reader
		if payload != nil {
			reader = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, reqURL, reader)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("bitculator: building request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.userAgent)
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.hc.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, nil, nil, ctx.Err()
			}
			if isIdempotent && attempt < c.maxRetries {
				if werr := sleep(ctx, backoff(attempt)); werr != nil {
					return nil, nil, nil, werr
				}
				continue
			}
			if isTimeout(err) {
				return nil, nil, nil, &TimeoutError{Path: path, Err: err}
			}
			return nil, nil, nil, &ConnectionError{Path: path, Err: err}
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		c.setQuota(parseQuota(resp.Header))

		if readErr != nil {
			if isIdempotent && attempt < c.maxRetries {
				if werr := sleep(ctx, backoff(attempt)); werr != nil {
					return nil, nil, nil, werr
				}
				continue
			}
			return nil, nil, nil, &ConnectionError{Path: path, Err: readErr}
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			var env envelope
			if len(respBody) > 0 {
				if err := json.Unmarshal(respBody, &env); err != nil {
					// Genuinely malformed JSON is a decode error; valid JSON that
					// simply isn't envelope-shaped (e.g. openapi.json) passes
					// through raw.
					if !json.Valid(respBody) {
						return nil, nil, nil, &DecodeError{Path: path, Err: err}
					}
					return json.RawMessage(respBody), nil, nil, nil
				}
			}
			data := env.Data
			if data == nil {
				data = json.RawMessage(respBody)
			}
			return data, env.Meta, env.Links, nil
		}

		if (resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode >= 500 && isIdempotent)) && attempt < c.maxRetries {
			if werr := sleep(ctx, retryDelay(resp.Header, attempt)); werr != nil {
				return nil, nil, nil, werr
			}
			continue
		}
		return nil, nil, nil, newAPIError(resp.StatusCode, respBody, resp.Header)
	}
}

func (c *httpClient) setQuota(q *Quota) {
	c.mu.Lock()
	c.quota = q
	c.mu.Unlock()
}

func (c *httpClient) getQuota() *Quota {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.quota
}

// ── helpers (all pure; the single place these concerns are implemented) ──────

// enc percent-encodes a dynamic path segment (slug or id) so it is URL-safe.
func enc(v any) string {
	return url.PathEscape(fmt.Sprint(v))
}

// encodeParams serializes params: drop nil, slices→CSV, floats as plain
// decimals, everything else via Sprint.
func encodeParams(params Params) string {
	if len(params) == 0 {
		return ""
	}
	values := url.Values{}
	for key, v := range params {
		if v == nil {
			continue
		}
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			if rv.Len() == 0 {
				continue
			}
			parts := make([]string, rv.Len())
			for i := range rv.Len() {
				parts[i] = paramString(rv.Index(i).Interface())
			}
			values.Set(key, strings.Join(parts, ","))
		default:
			values.Set(key, paramString(v))
		}
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}

// paramString stringifies a single param value. Floats are formatted as plain
// decimals (never e-notation) so the wire format matches the rest of the SDK
// family.
func paramString(v any) string {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}

func parseQuota(h http.Header) *Quota {
	raw := map[string]string{}
	for key, vals := range h {
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "x-quota-") && len(vals) > 0 {
			raw[lower] = vals[0]
		}
	}
	num := func(name string) *int {
		if v, ok := raw[name]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				return &n
			}
		}
		return nil
	}
	return &Quota{Limit: num("x-quota-limit"), Remaining: num("x-quota-remaining"), Reset: num("x-quota-reset"), Raw: raw}
}

func retryDelay(h http.Header, attempt int) time.Duration {
	if secs, err := strconv.Atoi(h.Get("Retry-After")); err == nil && secs > 0 {
		// Cap the honored Retry-After at 60s: a misbehaving proxy must not
		// block a synchronous call for an hour.
		return time.Duration(min(secs, 60)) * time.Second
	}
	return backoff(attempt)
}

// backoff is exponential with jitter, capped at 8s.
func backoff(attempt int) time.Duration {
	ms := min(250*(1<<min(attempt, 20)), 8000)
	return time.Duration(ms+rand.IntN(100)) * time.Millisecond
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func isTimeout(err error) bool {
	var timeouter interface{ Timeout() bool }
	if errors.As(err, &timeouter) {
		return timeouter.Timeout()
	}
	return errors.Is(err, context.DeadlineExceeded)
}
