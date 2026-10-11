package bitculator

import (
	"context"
	"encoding/json"
)

// GlobalMarket covers global market aggregates.
type GlobalMarket struct{ c *httpClient }

// Snapshot returns the global market snapshot (total marketcap, volume, dominance, ...).
func (r *GlobalMarket) Snapshot(ctx context.Context) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/global", nil)
}

// Heatmap returns market heatmap data.
func (r *GlobalMarket) Heatmap(ctx context.Context) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/global/heatmap", nil)
}

// History returns global marketcap or volume history. metric is e.g. "marketcap"
// or "volume". Params: period.
func (r *GlobalMarket) History(ctx context.Context, metric string, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/global/history/"+enc(metric), params)
}

// Sentiment covers community sentiment and market-wide sentiment indices.
type Sentiment struct{ c *httpClient }

// Votes returns the community vote tallies for a coin.
func (r *Sentiment) Votes(ctx context.Context, slug string) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/coins/"+enc(slug)+"/votes", nil)
}

// Vote casts a sentiment vote for a coin ("bullish" or "bearish").
func (r *Sentiment) Vote(ctx context.Context, slug, vote string) (json.RawMessage, error) {
	return r.c.post(ctx, "/api/v1/coins/"+enc(slug)+"/votes", map[string]any{"vote": vote})
}

// FearGreed returns the Fear & Greed index. Params: coin. Each reading carries
// "index" (0-100, as shown on Bitculator; use it for display) and the raw
// "score" (-100..+100).
func (r *Sentiment) FearGreed(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/sentiment/fear-greed", params)
}

// BullBear returns the Bull / Bear index. Params: coin.
func (r *Sentiment) BullBear(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/sentiment/bull-bear", params)
}

// Altseason returns the Altseason index. Params: days.
func (r *Sentiment) Altseason(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/sentiment/altseason", params)
}

// Indicators returns the aggregate market indicator tally.
func (r *Sentiment) Indicators(ctx context.Context) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/sentiment/indicators", nil)
}

// Liquidations covers the liquidation feed and aggregates.
type Liquidations struct{ c *httpClient }

// Feed returns the live liquidation feed (paginated). Params: page, per_page,
// exchange, instrument, position, order, symbol, min_usd.
func (r *Liquidations) Feed(ctx context.Context, params Params) (*Page, error) {
	return r.c.page(ctx, "/api/v1/liquidations", params)
}

// Hourly returns hourly liquidation totals. Params: hours.
func (r *Liquidations) Hourly(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/liquidations/hourly", params)
}

// Daily returns daily liquidation totals. Params: days.
func (r *Liquidations) Daily(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/liquidations/daily", params)
}

// Summary returns today's liquidation summary.
func (r *Liquidations) Summary(ctx context.Context) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/liquidations/summary", nil)
}

// Netflow returns long/short liquidation netflow. Params: days.
func (r *Liquidations) Netflow(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/liquidations/netflow", params)
}

// Coins returns the most-liquidated coins. Params: hours, limit.
func (r *Liquidations) Coins(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/liquidations/coins", params)
}
