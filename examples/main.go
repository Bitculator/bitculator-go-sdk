// Run with: BITCULATOR_API_KEY=... go run ./examples
// (against local dev, add BITCULATOR_BASE_URL=http://localhost)
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	bitculator "github.com/bitculator/bitculator-go-sdk"
)

func main() {
	client := bitculator.New(
		os.Getenv("BITCULATOR_API_KEY"),
		bitculator.WithBaseURL(getenvDefault("BITCULATOR_BASE_URL", "https://bitculator.com")),
	)
	ctx := context.Background()

	if _, err := client.Meta.Ping(ctx); err != nil {
		fail(err)
	}

	// Single resource — decode the raw JSON into your own type.
	raw, err := client.Coins.Get(ctx, "bitcoin", nil)
	if err != nil {
		fail(err)
	}
	var coin struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"` // decimal string, never rounded
	}
	if err := json.Unmarshal(raw, &coin); err != nil {
		fail(err)
	}
	fmt.Println("BTC price:", coin.Price)

	// Auto-paginate the first ~40 gainers.
	page, err := client.Coins.Gainers(ctx, bitculator.Params{"per_page": 20})
	if err != nil {
		fail(err)
	}
	seen := 0
	for row, err := range page.All(ctx) {
		if err != nil {
			fail(err)
		}
		var g struct {
			Symbol string `json:"symbol"`
		}
		_ = json.Unmarshal(row, &g)
		fmt.Println("gainer:", g.Symbol)
		if seen++; seen >= 40 {
			break
		}
	}

	fmt.Printf("quota: %+v\n", client.Quota())
}

func getenvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func fail(err error) {
	var apiErr *bitculator.APIError
	if errors.As(err, &apiErr) {
		fmt.Fprintf(os.Stderr, "API %d: %s %s\n", apiErr.Status, apiErr.Code, apiErr.Message)
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(1)
}
