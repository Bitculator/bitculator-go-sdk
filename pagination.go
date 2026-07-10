package bitculator

import (
	"context"
	"encoding/json"
	"iter"
)

// PageMeta is the Laravel paginator metadata on every paginated list response.
type PageMeta struct {
	CurrentPage int  `json:"current_page"`
	From        *int `json:"from"`
	LastPage    int  `json:"last_page"`
	PerPage     int  `json:"per_page"`
	To          *int `json:"to"`
	Total       int  `json:"total"`
}

// PageLinks are the navigation links on every paginated list response.
type PageLinks struct {
	First *string `json:"first"`
	Last  *string `json:"last"`
	Prev  *string `json:"prev"`
	Next  *string `json:"next"`
}

// Page is one page of a paginated list, plus lazy forward navigation. Rows are
// raw JSON; decode them with Into or json.Unmarshal.
type Page struct {
	Data  []json.RawMessage
	Meta  PageMeta
	Links PageLinks

	c      *httpClient
	path   string
	params Params
}

// HasMore reports whether at least one more page exists.
func (p *Page) HasMore() bool {
	if p.Links.Next != nil && *p.Links.Next != "" {
		return true
	}
	return p.Meta.CurrentPage != 0 && p.Meta.LastPage != 0 && p.Meta.CurrentPage < p.Meta.LastPage
}

// Next fetches the next page, or returns (nil, nil) if this is the last one.
func (p *Page) Next(ctx context.Context) (*Page, error) {
	if !p.HasMore() {
		return nil, nil
	}
	params := Params{}
	for k, v := range p.params {
		params[k] = v
	}
	params["page"] = p.Meta.CurrentPage + 1
	return p.c.page(ctx, p.path, params)
}

// All returns an iterator over every row across this and all following pages,
// fetching lazily:
//
//	for row, err := range page.All(ctx) {
//	    if err != nil { ... }
//	    var coin Coin
//	    json.Unmarshal(row, &coin)
//	}
func (p *Page) All(ctx context.Context) iter.Seq2[json.RawMessage, error] {
	return func(yield func(json.RawMessage, error) bool) {
		page := p
		for page != nil {
			for _, row := range page.Data {
				if !yield(row, nil) {
					return
				}
			}
			next, err := page.Next(ctx)
			if err != nil {
				yield(nil, err)
				return
			}
			page = next
		}
	}
}

// Into unmarshals the current page's rows into dst (a pointer to a slice).
func (p *Page) Into(dst any) error {
	arr, err := json.Marshal(p.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(arr, dst)
}
