package bitculator

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestSendsBearerAuthAcceptAndUserAgent(t *testing.T) {
	var seen http.Header
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Clone()
		respond(w, 200, `{"data":{}}`, nil)
	}))
	if _, err := client.Coins.Get(context.Background(), "bitcoin", nil); err != nil {
		t.Fatal(err)
	}
	if got := seen.Get("Authorization"); got != "Bearer test-key" {
		t.Fatalf("Authorization = %q", got)
	}
	if got := seen.Get("Accept"); got != "application/json" {
		t.Fatalf("Accept = %q", got)
	}
	if got := seen.Get("User-Agent"); got != "bitculator-go/test" {
		t.Fatalf("User-Agent = %q", got)
	}
	if got := seen.Get("Content-Type"); got != "" {
		t.Fatalf("GET should send no Content-Type, got %q", got)
	}
}

func TestPostSendsAJSONBody(t *testing.T) {
	var seen http.Header
	var body []byte
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Clone()
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		body = buf
		respond(w, 200, `{"data":{}}`, nil)
	}))
	if _, err := client.http.post(context.Background(), "/api/v1/alarms", map[string]any{"asset_slug": "bitcoin"}); err != nil {
		t.Fatal(err)
	}
	if got := seen.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
	if !strings.Contains(string(body), `"asset_slug":"bitcoin"`) {
		t.Fatalf("body = %s", body)
	}
}

func TestUnencodableBodyFailsWithoutARequest(t *testing.T) {
	hits := 0
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		respond(w, 200, `{"data":{}}`, nil)
	}))
	_, err := client.http.post(context.Background(), "/api/v1/alarms", make(chan int))
	if err == nil || !strings.Contains(err.Error(), "encoding request body") {
		t.Fatalf("err = %v", err)
	}
	if hits != 0 {
		t.Fatalf("request went out despite encode failure")
	}
}

func TestUnwrapsTheDataEnvelope(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(w, 200, `{"data":{"name":"Bitcoin","price":"67412.83451201"}}`, nil)
	}))
	raw, err := client.Coins.Get(context.Background(), "bitcoin", nil)
	if err != nil {
		t.Fatal(err)
	}
	var coin struct {
		Name  string `json:"name"`
		Price string `json:"price"`
	}
	if err := json.Unmarshal(raw, &coin); err != nil {
		t.Fatal(err)
	}
	if coin.Name != "Bitcoin" || coin.Price != "67412.83451201" {
		t.Fatalf("coin = %+v", coin)
	}
}

func TestPassesThroughNonEnvelopeObjects(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(w, 200, `{"openapi":"3.0.0"}`, nil)
	}))
	raw, err := client.http.get(context.Background(), "/api/v1/openapi.json", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"openapi"`) {
		t.Fatalf("raw = %s", raw)
	}
}

func TestPassesThroughValidNonObjectJSON(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(w, 200, `[1,2,3]`, nil)
	}))
	raw, err := client.http.get(context.Background(), "/api/v1/whatever", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `[1,2,3]` {
		t.Fatalf("raw = %s", raw)
	}
}

func TestEmpty2xxBodyIsANilResult(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	raw, err := client.http.delete(context.Background(), "/api/v1/alarms/1")
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 0 {
		t.Fatalf("raw = %s", raw)
	}
}

func TestMalformed2xxBodyIsADecodeError(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(w, 200, `<html>gateway page</html>`, nil)
	}))
	_, err := client.Coins.Get(context.Background(), "bitcoin", nil)
	var decodeErr *DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("err = %v", err)
	}
}

func TestParsesQuotaHeaders(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(w, 200, `{"data":[]}`, map[string]string{
			"X-Quota-Limit":     "10000",
			"X-Quota-Remaining": "9871",
			"X-Quota-Reset":     "1720569600",
		})
	}))
	if quota := client.Quota(); quota != nil {
		t.Fatalf("quota before first call = %+v", quota)
	}
	if _, err := client.Coins.Get(context.Background(), "bitcoin", nil); err != nil {
		t.Fatal(err)
	}
	quota := client.Quota()
	if quota == nil || *quota.Limit != 10000 || *quota.Remaining != 9871 || *quota.Reset != 1720569600 {
		t.Fatalf("quota = %+v", quota)
	}
	if quota.Raw["x-quota-limit"] != "10000" {
		t.Fatalf("raw = %v", quota.Raw)
	}
}

func TestMissingQuotaHeadersParseToNil(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(w, 200, `{"data":[]}`, nil)
	}))
	if _, err := client.Coins.Get(context.Background(), "bitcoin", nil); err != nil {
		t.Fatal(err)
	}
	quota := client.Quota()
	if quota == nil || quota.Limit != nil || quota.Remaining != nil || quota.Reset != nil || len(quota.Raw) != 0 {
		t.Fatalf("quota = %+v", quota)
	}
}
