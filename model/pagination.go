package model

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	DefaultLimit = 50
	MaxLimit     = 200
)

type PaginatedResponse struct {
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
	Data   interface{} `json:"data"`
}

type PaginationParams struct {
	Limit  int
	Offset int
}

func ParsePagination(r *http.Request) PaginationParams {
	p := PaginationParams{
		Limit:  DefaultLimit,
		Offset: 0,
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			p.Limit = n
		}
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}

	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			p.Offset = n
		}
	}

	return p
}

type SortParams struct {
	Field string
	Order string // "asc" or "desc"
}

func ParseSort(r *http.Request, validFields []string, defaultField string) SortParams {
	field := r.URL.Query().Get("sort")
	order := strings.ToLower(r.URL.Query().Get("order"))

	valid := false
	for _, f := range validFields {
		if field == f {
			valid = true
			break
		}
	}
	if !valid {
		field = defaultField
	}
	if order != "desc" {
		order = "asc"
	}
	return SortParams{Field: field, Order: order}
}
