package model

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePagination_Defaults(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	p := ParsePagination(r)
	if p.Limit != DefaultLimit {
		t.Errorf("default limit = %d, want %d", p.Limit, DefaultLimit)
	}
	if p.Offset != 0 {
		t.Errorf("default offset = %d, want 0", p.Offset)
	}
}

func TestParsePagination_Custom(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?limit=20&offset=100", nil)
	p := ParsePagination(r)
	if p.Limit != 20 {
		t.Errorf("limit = %d, want 20", p.Limit)
	}
	if p.Offset != 100 {
		t.Errorf("offset = %d, want 100", p.Offset)
	}
}

func TestParsePagination_MaxLimit(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?limit=999", nil)
	p := ParsePagination(r)
	if p.Limit != MaxLimit {
		t.Errorf("limit = %d, want %d (capped)", p.Limit, MaxLimit)
	}
}

func TestParsePagination_InvalidValues(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedLimit  int
		expectedOffset int
	}{
		{"negative limit", "limit=-5", DefaultLimit, 0},
		{"zero limit", "limit=0", DefaultLimit, 0},
		{"non-numeric limit", "limit=abc", DefaultLimit, 0},
		{"negative offset", "offset=-1", DefaultLimit, 0},
		{"non-numeric offset", "offset=xyz", DefaultLimit, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)
			p := ParsePagination(r)
			if p.Limit != tt.expectedLimit {
				t.Errorf("limit = %d, want %d", p.Limit, tt.expectedLimit)
			}
			if p.Offset != tt.expectedOffset {
				t.Errorf("offset = %d, want %d", p.Offset, tt.expectedOffset)
			}
		})
	}
}
