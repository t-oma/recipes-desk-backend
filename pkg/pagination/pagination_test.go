package pagination_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"recipes-desk/pkg/pagination"
)

func TestRequest_Skip(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		limit    int
		expected int64
	}{
		{
			name:     "page 1 returns skip 0",
			page:     1,
			limit:    20,
			expected: 0,
		},
		{
			name:     "page 2 with limit 10 returns skip 10",
			page:     2,
			limit:    10,
			expected: 10,
		},
		{
			name:     "page 3 with limit 20 returns skip 40",
			page:     3,
			limit:    20,
			expected: 40,
		},
		{
			name:     "page 5 with limit 5 returns skip 20",
			page:     5,
			limit:    5,
			expected: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pagination.Request{
				Page:  tt.page,
				Limit: tt.limit,
			}
			assert.Equal(t, tt.expected, req.Skip())
		})
	}
}

func TestNewResult_Metadata(t *testing.T) {
	tests := []struct {
		name            string
		items           []string
		req             *pagination.Request
		total           int64
		expectedPages   int
		expectedHasNext bool
		expectedHasPrev bool
	}{
		{
			name:            "first page with more items",
			items:           []string{"item1", "item2", "item3"},
			req:             &pagination.Request{Page: 1, Limit: 10},
			total:           25,
			expectedPages:   3,
			expectedHasNext: true,
			expectedHasPrev: false,
		},
		{
			name:            "middle page",
			items:           []string{"item11", "item12"},
			req:             &pagination.Request{Page: 2, Limit: 10},
			total:           25,
			expectedPages:   3,
			expectedHasNext: true,
			expectedHasPrev: true,
		},
		{
			name:            "last page",
			items:           []string{"item21", "item22", "item23", "item24", "item25"},
			req:             &pagination.Request{Page: 3, Limit: 10},
			total:           25,
			expectedPages:   3,
			expectedHasNext: false,
			expectedHasPrev: true,
		},
		{
			name:            "single page exact match",
			items:           []string{"item1", "item2", "item3"},
			req:             &pagination.Request{Page: 1, Limit: 3},
			total:           3,
			expectedPages:   1,
			expectedHasNext: false,
			expectedHasPrev: false,
		},
		{
			name:            "empty result",
			items:           []string{},
			req:             &pagination.Request{Page: 1, Limit: 10},
			total:           0,
			expectedPages:   0,
			expectedHasNext: false,
			expectedHasPrev: false,
		},
		{
			name:            "page beyond total",
			items:           []string{},
			req:             &pagination.Request{Page: 10, Limit: 10},
			total:           5,
			expectedPages:   1,
			expectedHasNext: false,
			expectedHasPrev: true,
		},
		{
			name:            "limit below 1",
			items:           []string{},
			req:             &pagination.Request{Page: 1, Limit: 0},
			total:           5,
			expectedPages:   5,
			expectedHasNext: true,
			expectedHasPrev: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pagination.NewResult(tt.items, tt.req, tt.total)

			assert.Equal(t, tt.req.Page, result.Pagination.Page)
			assert.Equal(t, tt.req.Limit, result.Pagination.Limit)
			assert.Equal(t, tt.total, result.Pagination.Total)
			assert.Equal(t, tt.expectedPages, result.Pagination.TotalPages)
			assert.Equal(t, tt.expectedHasNext, result.Pagination.HasNext())
			assert.Equal(t, tt.expectedHasNext, result.HasNext())
			assert.Equal(t, tt.expectedHasPrev, result.Pagination.HasPrev())
			assert.Equal(t, tt.expectedHasPrev, result.HasPrev())
			assert.Len(t, tt.items, len(result.Items))
		})
	}
}

func TestNewResult_TotalPages_Calculation(t *testing.T) {
	tests := []struct {
		total         int64
		limit         int
		expectedPages int
	}{
		{total: 0, limit: 10, expectedPages: 0},
		{total: 5, limit: 10, expectedPages: 1},
		{total: 10, limit: 10, expectedPages: 1},
		{total: 11, limit: 10, expectedPages: 2},
		{total: 20, limit: 10, expectedPages: 2},
		{total: 21, limit: 10, expectedPages: 3},
		{total: 100, limit: 20, expectedPages: 5},
		{total: 101, limit: 20, expectedPages: 6},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("total_%d_limit_%d", tt.total, tt.limit), func(t *testing.T) {
			req := &pagination.Request{Page: 1, Limit: tt.limit}
			result := pagination.NewResult([]string{}, req, tt.total)
			assert.Equal(t, tt.expectedPages, result.Pagination.TotalPages)
		})
	}
}

func TestMetadata_HasNext(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		totalPages int
		expected   bool
	}{
		{name: "page 1 of 3", page: 1, totalPages: 3, expected: true},
		{name: "page 2 of 3", page: 2, totalPages: 3, expected: true},
		{name: "page 3 of 3", page: 3, totalPages: 3, expected: false},
		{name: "page 1 of 1", page: 1, totalPages: 1, expected: false},
		{name: "page 5 of 3", page: 5, totalPages: 3, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := pagination.Metadata{ //nolint:exhaustruct // test struct
				Page:       tt.page,
				TotalPages: tt.totalPages,
			}
			assert.Equal(t, tt.expected, meta.HasNext())
		})
	}
}

func TestMetadata_HasPrev(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		expected bool
	}{
		{"page 1", 1, false},
		{"page 2", 2, true},
		{"page 5", 5, true},
		{"page 0", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := pagination.Metadata{ //nolint:exhaustruct // test struct
				Page: tt.page,
			}
			assert.Equal(t, tt.expected, meta.HasPrev())
		})
	}
}
