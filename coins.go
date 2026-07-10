package bitculator

import (
	"context"
	"encoding/json"
)

// Coins covers coins & tokens: listings, movers, detail, and history.
type Coins struct{ c *httpClient }

// List returns coins (paginated). Params: page, per_page, type, status, search,
// min_price, max_price, min_marketcap, max_marketcap, min_volume, max_volume,
// ids, slugs, symbols, sort, interval.
func (r *Coins) List(ctx context.Context, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/coins", params)
}

// RecentlyAdded returns recently added coins (paginated).
func (r *Coins) RecentlyAdded(ctx context.Context, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/coins/recently-added", params)
}

// Gainers returns the top gainers (paginated).
func (r *Coins) Gainers(ctx context.Context, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/coins/gainers", params)
}

// Losers returns the top losers (paginated).
func (r *Coins) Losers(ctx context.Context, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/coins/losers", params)
}

// Trending returns trending coins. Params: limit.
func (r *Coins) Trending(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/coins/trending", params)
}

// Get returns a single coin by slug. Params: locale.
func (r *Coins) Get(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/coins/"+enc(slug), params)
}

// History returns OHLC candle history. Params: interval, start, end, limit.
func (r *Coins) History(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/coins/"+enc(slug)+"/history", params)
}

// MarketcapHistory returns marketcap history. Params: interval, start, end, limit.
func (r *Coins) MarketcapHistory(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/coins/"+enc(slug)+"/marketcap-history", params)
}

// Sparkline returns a price sparkline. Params: period.
func (r *Coins) Sparkline(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/coins/"+enc(slug)+"/sparkline", params)
}

// Prices covers lightweight price lookups.
type Prices struct{ c *httpClient }

// List returns prices for many coins at once. Params: ids, slugs, symbols, convert.
func (r *Prices) List(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/prices", params)
}

// Get returns the price of a single coin. Params: convert.
func (r *Prices) Get(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/prices/"+enc(slug), params)
}

// Historical returns the historical price of a coin on a date. Params: slug, date.
func (r *Prices) Historical(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/historical-price", params)
}

// Conversion covers asset conversion and reference rates.
type Conversion struct{ c *httpClient }

// Convert converts an amount between two assets. Params: from, to, amount.
func (r *Conversion) Convert(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/convert", params)
}

// Fiats lists supported fiat currencies.
func (r *Conversion) Fiats(ctx context.Context) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/fiats", nil)
}

// Rates lists conversion rates.
func (r *Conversion) Rates(ctx context.Context) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/rates", nil)
}
