package bitculator

import (
	"context"
	"encoding/json"
)

// Indicators covers technical indicators for a coin.
//
// Period indicators (RSI, StochRSI, SMA, CCI, MFI, WilliamsR, PriceVolatility,
// VolumeVolatility) accept: period, start, end, limit.
// Series indicators (MACD, OBV, ADX, VWAP, CMF) accept: start, end, limit.
type Indicators struct{ c *httpClient }

func (r *Indicators) indicator(ctx context.Context, slug, name string, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/coins/"+enc(slug)+"/indicators/"+name, params)
}

// Snapshot returns all indicators for a coin at a glance.
func (r *Indicators) Snapshot(ctx context.Context, slug string) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/coins/"+enc(slug)+"/indicators", nil)
}

// RSI returns the Relative Strength Index.
func (r *Indicators) RSI(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.indicator(ctx, slug, "rsi", params)
}

// StochRSI returns the Stochastic RSI.
func (r *Indicators) StochRSI(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.indicator(ctx, slug, "stoch-rsi", params)
}

// SMA returns the Simple Moving Average.
func (r *Indicators) SMA(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.indicator(ctx, slug, "sma", params)
}

// CCI returns the Commodity Channel Index.
func (r *Indicators) CCI(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.indicator(ctx, slug, "cci", params)
}

// MFI returns the Money Flow Index.
func (r *Indicators) MFI(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.indicator(ctx, slug, "mfi", params)
}

// WilliamsR returns Williams %R.
func (r *Indicators) WilliamsR(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.indicator(ctx, slug, "williams-r", params)
}

// PriceVolatility returns price volatility.
func (r *Indicators) PriceVolatility(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.indicator(ctx, slug, "price-volatility", params)
}

// VolumeVolatility returns volume volatility.
func (r *Indicators) VolumeVolatility(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.indicator(ctx, slug, "volume-volatility", params)
}

// MACD returns the Moving Average Convergence Divergence.
func (r *Indicators) MACD(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.indicator(ctx, slug, "macd", params)
}

// OBV returns On-Balance Volume.
func (r *Indicators) OBV(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.indicator(ctx, slug, "obv", params)
}

// ADX returns the Average Directional Index.
func (r *Indicators) ADX(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.indicator(ctx, slug, "adx", params)
}

// VWAP returns the Volume-Weighted Average Price.
func (r *Indicators) VWAP(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.indicator(ctx, slug, "vwap", params)
}

// CMF returns Chaikin Money Flow.
func (r *Indicators) CMF(ctx context.Context, slug string, params Params) (json.RawMessage, error) {
	return r.indicator(ctx, slug, "cmf", params)
}

// Calculators covers financial calculators.
type Calculators struct{ c *httpClient }

// DCA runs the dollar-cost-averaging calculator. Params: slug, amount, interval,
// start, end, series.
func (r *Calculators) DCA(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/calculators/dca", params)
}

// ProfitLoss runs the profit / loss calculator. Params: slug, amount, buy_date,
// sell_date, buy_fee, sell_fee.
func (r *Calculators) ProfitLoss(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/calculators/profit-loss", params)
}

// CompoundInterest runs the compound interest calculator. Params: principal,
// rate, duration, duration_unit, compound_frequency, contribution,
// contribution_frequency. "rate" is the annual nominal rate in %; each
// compounding period earns "rate" divided by the periods per year.
func (r *Calculators) CompoundInterest(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/calculators/compound-interest", params)
}

// Loan runs the loan-vs-sell calculator. Params: slug, crypto_amount,
// needed_cash, term_months, interest_rate, ltv, expected_growth, tax_rate.
func (r *Calculators) Loan(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/calculators/loan", params)
}

// Staking runs the staking rewards calculator. Params: amount, period,
// period_unit, apy, compound_frequency, commission.
func (r *Calculators) Staking(ctx context.Context, params Params) (json.RawMessage, error) {
	return r.c.get(ctx, "/api/v1/calculators/staking", params)
}
