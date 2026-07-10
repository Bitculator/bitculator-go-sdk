package bitculator

import (
	"net/url"
	"testing"
)

func parsed(t *testing.T, params Params) url.Values {
	t.Helper()
	raw := encodeParams(params)
	if raw == "" {
		return url.Values{}
	}
	values, err := url.ParseQuery(raw[1:])
	if err != nil {
		t.Fatalf("unparsable query %q: %v", raw, err)
	}
	return values
}

func TestSlicesSerializeToCSV(t *testing.T) {
	values := parsed(t, Params{"symbols": []string{"btc", "eth", "sol"}})
	if got := values.Get("symbols"); got != "btc,eth,sol" {
		t.Fatalf("symbols = %q", got)
	}
}

func TestNilAndEmptySlicesAreDropped(t *testing.T) {
	values := parsed(t, Params{"locale": nil, "symbols": []string{}, "page": 1})
	if values.Has("locale") || values.Has("symbols") {
		t.Fatalf("nil/empty params leaked: %v", values)
	}
	if got := values.Get("page"); got != "1" {
		t.Fatalf("page = %q", got)
	}
}

func TestBooleansSerializeLowercase(t *testing.T) {
	values := parsed(t, Params{"include_inactive": true, "compact": false})
	if values.Get("include_inactive") != "true" || values.Get("compact") != "false" {
		t.Fatalf("bools = %v", values)
	}
}

func TestTinyFloatsExpandToPlainDecimals(t *testing.T) {
	values := parsed(t, Params{"min_price": 1e-7, "max_price": -2.5e-8})
	if got := values.Get("min_price"); got != "0.0000001" {
		t.Fatalf("min_price = %q", got)
	}
	if got := values.Get("max_price"); got != "-0.000000025" {
		t.Fatalf("max_price = %q", got)
	}
}

func TestHugeFloatsExpandToPlainDecimals(t *testing.T) {
	values := parsed(t, Params{"min_marketcap": 1e21})
	if got := values.Get("min_marketcap"); got != "1000000000000000000000" {
		t.Fatalf("min_marketcap = %q", got)
	}
}

func TestFloatsInsideSlicesExpandToo(t *testing.T) {
	values := parsed(t, Params{"prices": []float64{1e-7, 0.5}})
	if got := values.Get("prices"); got != "0.0000001,0.5" {
		t.Fatalf("prices = %q", got)
	}
}

func TestOrdinaryNumbersKeepTheirShortestForm(t *testing.T) {
	values := parsed(t, Params{"min_price": 0.1, "page": 42})
	if values.Get("min_price") != "0.1" || values.Get("page") != "42" {
		t.Fatalf("values = %v", values)
	}
}

func TestNoParamsMeansNoQueryString(t *testing.T) {
	if got := encodeParams(nil); got != "" {
		t.Fatalf("nil params = %q", got)
	}
	if got := encodeParams(Params{"only": nil}); got != "" {
		t.Fatalf("all-nil params = %q", got)
	}
}

func TestEncPercentEncodesPathSegments(t *testing.T) {
	if got := enc("weird/slug name"); got != "weird%2Fslug%20name" {
		t.Fatalf("enc = %q", got)
	}
	if got := enc(42); got != "42" {
		t.Fatalf("enc int = %q", got)
	}
}
