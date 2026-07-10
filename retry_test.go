package bitculator

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

// seqHandler serves the nth queued status on the nth request (the last repeats)
// and counts hits. 2xx statuses reply {"data":"ok"}.
func seqHandler(hits *atomic.Int32, statuses ...int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(hits.Add(1))
		status := statuses[min(n, len(statuses))-1]
		if status >= 200 && status < 300 {
			respond(w, status, `{"data":"ok"}`, nil)
			return
		}
		respond(w, status, errorBody("err", "nope"), nil)
	})
}

func TestGetRetriesA5xxAndSucceeds(t *testing.T) {
	var hits atomic.Int32
	client := newTestClient(t, seqHandler(&hits, 500, 200))
	if _, err := client.http.get(context.Background(), "/api/v1/coins", nil); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Fatalf("hits = %d", hits.Load())
	}
}

func TestDeleteRetriesA5xxAndSucceeds(t *testing.T) {
	var hits atomic.Int32
	client := newTestClient(t, seqHandler(&hits, 500, 200))
	if _, err := client.http.delete(context.Background(), "/api/v1/alarms/1"); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Fatalf("hits = %d", hits.Load())
	}
}

func TestPostNeverRetriesA5xx(t *testing.T) {
	var hits atomic.Int32
	client := newTestClient(t, seqHandler(&hits, 500, 200))
	_, err := client.http.post(context.Background(), "/api/v1/alarms", map[string]any{"x": 1})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !apiErr.IsServer() {
		t.Fatalf("err = %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d — a POST must never be re-sent on a 5xx", hits.Load())
	}
}

func TestPostDoesRetryA429(t *testing.T) {
	var hits atomic.Int32
	client := newTestClient(t, seqHandler(&hits, 429, 200))
	if _, err := client.http.post(context.Background(), "/api/v1/alarms", map[string]any{"x": 1}); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Fatalf("hits = %d", hits.Load())
	}
}

func TestGivesUpAfterMaxRetries(t *testing.T) {
	var hits atomic.Int32
	client := newTestClient(t, seqHandler(&hits, 500, 500, 500, 200), WithMaxRetries(2))
	_, err := client.http.get(context.Background(), "/api/v1/coins", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !apiErr.IsServer() {
		t.Fatalf("err = %v", err)
	}
	if hits.Load() != 3 { // initial attempt + 2 retries; the queued 200 is never reached
		t.Fatalf("hits = %d", hits.Load())
	}
}

func TestHonorsRetryAfterSecondsOnA429(t *testing.T) {
	var hits atomic.Int32
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			respond(w, 429, errorBody("rate_limited", "Slow down"), map[string]string{"Retry-After": "1"})
			return
		}
		respond(w, 200, `{"data":"ok"}`, nil)
	}))
	started := time.Now()
	if _, err := client.http.get(context.Background(), "/api/v1/coins", nil); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(started); elapsed < 900*time.Millisecond {
		t.Fatalf("elapsed = %v — Retry-After was not honored", elapsed)
	}
	if hits.Load() != 2 {
		t.Fatalf("hits = %d", hits.Load())
	}
}

func TestGetRetriesANetworkErrorThenConnects(t *testing.T) {
	// Nothing listens on the discard port — the retry gives the request a second life.
	client := New("test-key", WithBaseURL("http://127.0.0.1:9"), WithTimeout(500*time.Millisecond), WithMaxRetries(1))
	_, err := client.http.get(context.Background(), "/api/v1/coins", nil)
	var connErr *ConnectionError
	var timeoutErr *TimeoutError
	if !errors.As(err, &connErr) && !errors.As(err, &timeoutErr) {
		t.Fatalf("err = %v", err)
	}
}

func TestTimedOutGetRetriesThenThrows(t *testing.T) {
	var hits atomic.Int32
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		time.Sleep(400 * time.Millisecond)
		respond(w, 200, `{"data":"ok"}`, nil)
	}), WithTimeout(50*time.Millisecond), WithMaxRetries(1))
	_, err := client.http.get(context.Background(), "/api/v1/coins", nil)
	var timeoutErr *TimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("err = %v", err)
	}
	if hits.Load() != 2 { // the timeout is a transport failure, so the idempotent GET retried once
		t.Fatalf("hits = %d", hits.Load())
	}
}

func TestTimedOutPostFailsWithoutRetrying(t *testing.T) {
	var hits atomic.Int32
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		time.Sleep(400 * time.Millisecond)
		respond(w, 200, `{"data":"ok"}`, nil)
	}), WithTimeout(50*time.Millisecond))
	_, err := client.http.post(context.Background(), "/api/v1/alarms", map[string]any{"x": 1})
	var timeoutErr *TimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("err = %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d — a POST must never be re-sent on a timeout", hits.Load())
	}
}

func TestContextCancellationWinsOverRetrySleep(t *testing.T) {
	var hits atomic.Int32
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		respond(w, 429, errorBody("rate_limited", "Slow down"), map[string]string{"Retry-After": "30"})
	}))
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	started := time.Now()
	_, err := client.http.get(ctx, "/api/v1/coins", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
	if time.Since(started) > 5*time.Second {
		t.Fatal("cancellation did not interrupt the retry sleep")
	}
}
