# Bitculator Data API — Go SDK

[bitculator.com](https://bitculator.com) · [API documentation](https://bitculator.com/en/documentation/api/v1) · [Get an API key](https://bitculator.com/user/developer/api)
[pkg.go.dev](https://pkg.go.dev/github.com/bitculator/bitculator-go-sdk) · [GitHub](https://github.com/Bitculator/bitculator-go-sdk)

The official Go SDK for the [Bitculator Data API](https://bitculator.com).
Standard-library only (zero dependencies), context-aware, Go 1.23+.

> Keep your API key server-side. It is a Bearer token with the `data-api` ability
> and is not meant for client-side embedding.

## Install

```bash
go get github.com/bitculator/bitculator-go-sdk
```

## Quick start

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"

	bitculator "github.com/bitculator/bitculator-go-sdk"
)

func main() {
	client := bitculator.New("your-api-key")
	ctx := context.Background()

	// Single resource — the { "data": ... } envelope is unwrapped; rows are raw JSON.
	raw, _ := client.Coins.Get(ctx, "bitcoin", nil)
	var coin struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"` // decimal string, never rounded
	}
	json.Unmarshal(raw, &coin)
	fmt.Println(coin.Price) // "63520.780763913"

	// A paginated list.
	page, _ := client.Coins.List(ctx, bitculator.Params{"per_page": 50, "sort": "-marketcap"})
	fmt.Println(page.Meta.Total, len(page.Data))
}
```

## Prices are decimal strings

Prices, rates, and supplies come back as **strings** (e.g. `"63520.780763913"`)
to preserve full precision. Decode them into a `string` field (or a big-decimal
type) — never into a `float64`, which silently loses precision.

## Responses

Methods return raw `json.RawMessage` so you decode into your own types. Paginated
methods return a `*Page` whose `Data` is `[]json.RawMessage`; use `page.Into(&s)`
to decode the whole page, or decode rows one at a time.

## Pagination

```go
// One page at a time.
page, _ := client.Exchanges.List(ctx, bitculator.Params{"per_page": 100})
for page != nil {
	var exchanges []map[string]any
	page.Into(&exchanges)
	// ... use exchanges ...
	page, _ = page.Next(ctx)
}

// Or auto-paginate across every page (Go 1.23 range-over-func).
for row, err := range page.All(ctx) {
	if err != nil { /* handle */ }
	var coin map[string]any
	json.Unmarshal(row, &coin)
}
```

## Errors

HTTP failures return a `*APIError` carrying the `{ error: { code, message, details } }`
envelope; inspect it with `errors.As`.

```go
_, err := client.Coins.List(ctx, bitculator.Params{"per_page": 9999}) // over the plan cap -> 422
var apiErr *bitculator.APIError
if errors.As(err, &apiErr) {
	switch {
	case apiErr.IsValidation():  fmt.Println(apiErr.Code, string(apiErr.Details))
	case apiErr.IsRateLimited(): fmt.Println("retry after", apiErr.RetryAfter)
	default:                     fmt.Println(apiErr.Status, apiErr.Message)
	}
}
```

`IsAuth` (401) · `IsPermission` (403) · `IsNotFound` (404) · `IsValidation` (422) ·
`IsRateLimited` (429) · `IsServer` (5xx). Timeouts return `*TimeoutError`; network
failures return `*ConnectionError`.

## Quota

Every response carries `X-Quota-*` headers, surfaced on the client after each call:

```go
client.Coins.List(ctx, nil)
q := client.Quota() // *Quota { Limit, Remaining, Reset, Raw }
```

## Configuration

```go
client := bitculator.New("your-api-key",
	bitculator.WithBaseURL("https://bitculator.com"), // default
	bitculator.WithTimeout(30*time.Second),           // default
	bitculator.WithMaxRetries(2),                     // auto-retry 429/5xx/network; 0 to disable
	bitculator.WithHTTPClient(myClient),              // optional: custom *http.Client
)
```

## Resources

`Coins` · `Prices` · `Markets` · `Exchanges` · `Wallets` · `GlobalMarket` ·
`Sentiment` · `Indicators` · `Liquidations` · `Conversion` · `Calculators` ·
`Editorial` · `Alarms` · `Webhooks` · `Meta`

Full endpoint reference: <https://bitculator.com/en/documentation/api/v1>.

## License

MIT
