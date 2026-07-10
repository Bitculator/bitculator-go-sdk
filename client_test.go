package bitculator

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestExposesAll15ResourceGroups(t *testing.T) {
	client := New("k")
	groups := map[string]any{
		"Coins": client.Coins, "Prices": client.Prices, "Markets": client.Markets,
		"Exchanges": client.Exchanges, "Wallets": client.Wallets, "GlobalMarket": client.GlobalMarket,
		"Sentiment": client.Sentiment, "Indicators": client.Indicators, "Liquidations": client.Liquidations,
		"Conversion": client.Conversion, "Calculators": client.Calculators, "Editorial": client.Editorial,
		"Alarms": client.Alarms, "Webhooks": client.Webhooks, "Meta": client.Meta,
	}
	if len(groups) != 15 {
		t.Fatalf("expected 15 groups, listed %d", len(groups))
	}
	for name, group := range groups {
		if group == nil {
			t.Fatalf("client.%s is nil", name)
		}
	}
}

func TestResourceMethodsHitAPIV1Paths(t *testing.T) {
	var path string
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.EscapedPath()
		respond(w, 200, `{"data":{}}`, nil)
	}))
	if _, err := client.Coins.Get(context.Background(), "bitcoin", nil); err != nil {
		t.Fatal(err)
	}
	if path != "/api/v1/coins/bitcoin" {
		t.Fatalf("path = %q", path)
	}
}

func TestPercentEncodesDynamicPathSegments(t *testing.T) {
	var path string
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.EscapedPath()
		respond(w, 200, `{"data":{}}`, nil)
	}))
	if _, err := client.Coins.Get(context.Background(), "weird/slug name", nil); err != nil {
		t.Fatal(err)
	}
	if path != "/api/v1/coins/weird%2Fslug%20name" {
		t.Fatalf("path = %q", path)
	}
}

func TestWithBaseURLTrimsTrailingSlashes(t *testing.T) {
	var path string
	server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.EscapedPath()
		respond(w, 200, `{"data":{}}`, nil)
	}))
	// Rebuild against the same server URL with trailing slashes appended.
	base := server.http.baseURL + "///"
	client := New("k", WithBaseURL(base), WithUserAgent("bitculator-go/test"))
	if _, err := client.Coins.Get(context.Background(), "bitcoin", nil); err != nil {
		t.Fatal(err)
	}
	if path != "/api/v1/coins/bitcoin" {
		t.Fatalf("path = %q", path)
	}
}

func TestDefaultUserAgentCarriesTheVersion(t *testing.T) {
	var userAgent string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent = r.Header.Get("User-Agent")
		respond(w, 200, `{"data":{}}`, nil)
	})
	// Bypass newTestClient: it overrides the User-Agent.
	server := newTestClient(t, handler)
	client := New("k", WithBaseURL(server.http.baseURL))
	if _, err := client.Coins.Get(context.Background(), "bitcoin", nil); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(userAgent, "bitculator-go/") || !strings.Contains(userAgent, Version) {
		t.Fatalf("user agent = %q", userAgent)
	}
}
