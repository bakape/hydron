package common

import (
	"net/url"
	"strconv"
)

// Describes the contents a browser page is displaying
type Page struct {
	Page, PageTotal, Limit uint
	Order                  Order
	Filters                FilterSet
	Viewing                *Image // Image currently being viewed, if any
}

// Return the relative URL this page points to
func (p Page) URL() string {
	u := url.URL{
		Path:     "/search",
		RawQuery: p.Query(),
	}
	return u.String()
}

// Returns query string of page without leading '?'
func (p Page) Query() string {
	q := make(url.Values, 6)
	setUint := func(key string, i uint) {
		q.Set(key, strconv.FormatUint(uint64(i), 10))
	}

	setUint("page", p.Page)
	if p.Limit != 0 {
		setUint("limit", p.Limit)
	}
	if p.Order.Type != None {
		setUint("order", uint(p.Order.Type))
	}
	if p.Order.Reverse {
		q.Set("reverse", "on")
	}
	q.Set("q", p.Filters.String())
	if p.Viewing != nil {
		q.Set("img", p.Viewing.SHA1)
	}

	return q.Encode()
}

// Type of ordering for search results
type Order struct {
	Type    OrderType
	Reverse bool
}

// Types of ordering for search results
type OrderType uint8

const (
	None OrderType = iota
	BySize
	ByWidth
	ByHeight
	ByDuration
	ByTagCount
	Random
)
