package bitculator

import (
	"context"
	"encoding/json"
)

// Markets covers markets, tickers, and trading pairs.
type Markets struct{ c *httpClient }

// CoinMarkets returns markets (tickers) for a coin (paginated). Params: page,
// per_page, exchange, pair, instrument, search, min_volume, max_volume,
// min_change, max_change, sort.
func (r *Markets) CoinMarkets(ctx context.Context, slug string, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/coins/"+enc(slug)+"/markets", params)
}

// CoinSymbols returns the trading symbols for a coin.
func (r *Markets) CoinSymbols(ctx context.Context, slug string) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/coins/"+enc(slug)+"/symbols", nil)
}

// Tickers lists tickers across the market (paginated).
func (r *Markets) Tickers(ctx context.Context, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/tickers", params)
}

// Pairs lists trading pairs (paginated). Params: page, per_page, search, coin,
// min_volume, max_volume, sort.
func (r *Markets) Pairs(ctx context.Context, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/pairs", params)
}

// Pair returns a single trading pair by id.
func (r *Markets) Pair(ctx context.Context, id any) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/pairs/"+enc(id), nil)
}

// Exchanges covers exchange listings, trust score, history, markets, and assets.
type Exchanges struct{ c *httpClient }

// List returns exchanges (paginated). Params: page, per_page, type, search,
// min_pairs, max_pairs, min_assets, max_assets, min_volume, max_volume, ids,
// slugs, sort.
func (r *Exchanges) List(ctx context.Context, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/exchanges", params)
}

// Get returns a single exchange by slug.
func (r *Exchanges) Get(ctx context.Context, slug string) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/exchanges/"+enc(slug), nil)
}

// TrustScore returns an exchange's trust score.
func (r *Exchanges) TrustScore(ctx context.Context, slug string) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/exchanges/"+enc(slug)+"/trust-score", nil)
}

// History returns volume/history for an exchange. Params: interval, start, end, limit.
func (r *Exchanges) History(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/exchanges/"+enc(slug)+"/history", params)
}

// Sparkline returns a sparkline for an exchange. Params: period.
func (r *Exchanges) Sparkline(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/exchanges/"+enc(slug)+"/sparkline", params)
}

// Tickers returns markets (tickers) on an exchange (paginated).
func (r *Exchanges) Tickers(ctx context.Context, slug string, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/exchanges/"+enc(slug)+"/tickers", params)
}

// Assets returns coins listed on an exchange (paginated).
func (r *Exchanges) Assets(ctx context.Context, slug string, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/exchanges/"+enc(slug)+"/assets", params)
}

// Wallets covers wallet listings, release timeline, comparison, and supported coins.
type Wallets struct{ c *httpClient }

// List returns wallets (paginated). Params: page, per_page, search, min_score,
// max_score, tags, ids, slugs, sort.
func (r *Wallets) List(ctx context.Context, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/wallets", params)
}

// Timeline returns the wallet release timeline (paginated).
func (r *Wallets) Timeline(ctx context.Context, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/wallets/timeline", params)
}

// Compare compares a set of wallets side by side. Params: slugs.
func (r *Wallets) Compare(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/wallets/compare", params)
}

// Get returns a single wallet by slug.
func (r *Wallets) Get(ctx context.Context, slug string) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/wallets/"+enc(slug), nil)
}

// Assets returns coins supported by a wallet (paginated).
func (r *Wallets) Assets(ctx context.Context, slug string, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/wallets/"+enc(slug)+"/assets", params)
}
