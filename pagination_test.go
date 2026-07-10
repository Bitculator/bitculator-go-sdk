package bitculator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

type row struct {
	ID int `json:"id"`
}

// pageServer serves a deterministic 2-rows-per-page fixture and records queries.
func pageServer(hits *atomic.Int32, lastPage int, queries *[]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		*queries = append(*queries, r.URL.RawQuery)
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page == 0 {
			page = 1
		}
		first := (page - 1) * 2
		next := "null"
		if page < lastPage {
			next = fmt.Sprintf(`"/api/v1/coins?page=%d"`, page+1)
		}
		respond(w, 200, fmt.Sprintf(
			`{"data":[{"id":%d},{"id":%d}],"meta":{"current_page":%d,"last_page":%d,"per_page":2,"total":%d},"links":{"next":%s}}`,
			first+1, first+2, page, lastPage, lastPage*2, next,
		), nil)
	})
}

func TestPageExposesDataMetaAndLinks(t *testing.T) {
	var hits atomic.Int32
	var queries []string
	client := newTestClient(t, pageServer(&hits, 3, &queries))
	page, err := client.Coins.List(context.Background(), Params{"per_page": 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 2 || page.Meta.CurrentPage != 1 || page.Meta.LastPage != 3 {
		t.Fatalf("page = %+v", page)
	}
	if !page.HasMore() {
		t.Fatal("HasMore should be true on page 1 of 3")
	}
}

func TestNextFetchesTheFollowingPageAndPreservesTheQuery(t *testing.T) {
	var hits atomic.Int32
	var queries []string
	client := newTestClient(t, pageServer(&hits, 3, &queries))
	first, err := client.Coins.List(context.Background(), Params{"per_page": 2, "sort": "rank"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := first.Next(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var rows []row
	if err := second.Into(&rows); err != nil {
		t.Fatal(err)
	}
	if rows[0].ID != 3 || rows[1].ID != 4 {
		t.Fatalf("rows = %+v", rows)
	}
	q := queries[1]
	for _, want := range []string{"page=2", "per_page=2", "sort=rank"} {
		if !slices.Contains(strings.Split(q, "&"), want) {
			t.Fatalf("second query %q missing %q", q, want)
		}
	}
}

func TestNextReturnsNilOnTheLastPage(t *testing.T) {
	var hits atomic.Int32
	var queries []string
	client := newTestClient(t, pageServer(&hits, 1, &queries))
	page, err := client.Coins.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if page.HasMore() {
		t.Fatal("HasMore should be false on the last page")
	}
	next, err := page.Next(context.Background())
	if err != nil || next != nil {
		t.Fatalf("next = %v, err = %v", next, err)
	}
}

func TestAllWalksEveryPageInOrder(t *testing.T) {
	var hits atomic.Int32
	var queries []string
	client := newTestClient(t, pageServer(&hits, 3, &queries))
	page, err := client.Coins.List(context.Background(), Params{"per_page": 2})
	if err != nil {
		t.Fatal(err)
	}
	var seen []int
	for raw, err := range page.All(context.Background()) {
		if err != nil {
			t.Fatal(err)
		}
		var item row
		if err := json.Unmarshal(raw, &item); err != nil {
			t.Fatal(err)
		}
		seen = append(seen, item.ID)
	}
	if fmt.Sprint(seen) != "[1 2 3 4 5 6]" {
		t.Fatalf("seen = %v", seen)
	}
	if hits.Load() != 3 {
		t.Fatalf("hits = %d", hits.Load())
	}
}

func TestAllStopsWhenTheConsumerBreaks(t *testing.T) {
	var hits atomic.Int32
	var queries []string
	client := newTestClient(t, pageServer(&hits, 3, &queries))
	page, err := client.Coins.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, err := range page.All(context.Background()) {
		if err != nil {
			t.Fatal(err)
		}
		count++
		if count == 2 {
			break // stop inside page 1
		}
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d — breaking early must not fetch further pages", hits.Load())
	}
}

func TestEnvelopeWithoutMetaOrLinksIsASafeFinalPage(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(w, 200, `{"data":[{"id":9}]}`, nil)
	}))
	page, err := client.Coins.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 1 || page.HasMore() {
		t.Fatalf("page = %+v", page)
	}
	next, err := page.Next(context.Background())
	if err != nil || next != nil {
		t.Fatalf("next = %v, err = %v", next, err)
	}
}

func TestNonArrayRowsAreADecodeError(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(w, 200, `{"data":{"not":"an array"},"meta":{"current_page":1,"last_page":1}}`, nil)
	}))
	_, err := client.Coins.List(context.Background(), nil)
	var decodeErr *DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("err = %v", err)
	}
}
