// Package bitculator is the official Go SDK for the Bitculator Data API.
package bitculator

import (
	"net/http"
	"strings"
	"time"
)

// Version of the SDK, sent in the User-Agent header.
const Version = "1.1.0"

const defaultBaseURL = "https://bitculator.com"

// Client is the Bitculator Data API client. Create it with New. Each API group
// is a field: client.Coins, client.Prices, client.Exchanges, and so on.
type Client struct {
	Coins        *Coins
	Prices       *Prices
	Markets      *Markets
	Exchanges    *Exchanges
	Wallets      *Wallets
	GlobalMarket *GlobalMarket
	Sentiment    *Sentiment
	Indicators   *Indicators
	Liquidations *Liquidations
	Conversion   *Conversion
	Calculators  *Calculators
	Editorial    *Editorial
	Alarms       *Alarms
	Webhooks     *Webhooks
	Meta         *Meta

	http *httpClient
}

// Option configures the client.
type Option func(*httpClient)

// WithBaseURL overrides the API origin (default https://bitculator.com).
func WithBaseURL(u string) Option {
	return func(c *httpClient) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithTimeout sets the per-request timeout (default 30s).
func WithTimeout(d time.Duration) Option {
	return func(c *httpClient) { c.hc.Timeout = d }
}

// WithMaxRetries sets the number of automatic retries for 429/5xx/network errors (default 2).
func WithMaxRetries(n int) Option {
	return func(c *httpClient) { c.maxRetries = n }
}

// WithHTTPClient supplies a custom *http.Client (overrides WithTimeout).
func WithHTTPClient(h *http.Client) Option {
	return func(c *httpClient) { c.hc = h }
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(s string) Option {
	return func(c *httpClient) { c.userAgent = s }
}

// New creates a client with the given Data API key (a Bearer token with the
// data-api ability). Keep the key server-side.
func New(apiKey string, opts ...Option) *Client {
	hc := &httpClient{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		userAgent:  "bitculator-go/" + Version,
		maxRetries: 2,
		hc:         &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(hc)
	}

	return &Client{
		http:         hc,
		Coins:        &Coins{hc},
		Prices:       &Prices{hc},
		Markets:      &Markets{hc},
		Exchanges:    &Exchanges{hc},
		Wallets:      &Wallets{hc},
		GlobalMarket: &GlobalMarket{hc},
		Sentiment:    &Sentiment{hc},
		Indicators:   &Indicators{hc},
		Liquidations: &Liquidations{hc},
		Conversion:   &Conversion{hc},
		Calculators:  &Calculators{hc},
		Editorial:    &Editorial{hc},
		Alarms:       &Alarms{hc},
		Webhooks:     &Webhooks{hc},
		Meta:         &Meta{hc},
	}
}

// Quota returns the X-Quota-* values from the most recent response, or nil
// before the first call.
func (c *Client) Quota() *Quota {
	return c.http.getQuota()
}
