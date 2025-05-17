package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginationRequest_GetOffset(t *testing.T) {
	tests := []struct {
		name     string
		request  PaginationRequest
		expected int
	}{
		{
			name: "first page",
			request: PaginationRequest{
				Page:    1,
				PerPage: 10,
			},
			expected: 0,
		},
		{
			name: "second page",
			request: PaginationRequest{
				Page:    2,
				PerPage: 10,
			},
			expected: 10,
		},
		{
			name: "third page with different per page",
			request: PaginationRequest{
				Page:    3,
				PerPage: 15,
			},
			expected: 30,
		},
		{
			name:     "zero values",
			request:  PaginationRequest{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				actual := tt.request.GetOffset()
				assert.Equal(t, tt.expected, actual)
			},
		)
	}
}

func TestPaginationRequest_GetLimit(t *testing.T) {
	tests := []struct {
		name     string
		request  PaginationRequest
		expected int
	}{
		{
			name: "normal case",
			request: PaginationRequest{
				PerPage: 25,
			},
			expected: 25,
		},
		{
			name:     "zero value",
			request:  PaginationRequest{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				actual := tt.request.GetLimit()
				assert.Equal(t, tt.expected, actual)
			},
		)
	}
}

func TestPaginationRequest_GetPage(t *testing.T) {
	tests := []struct {
		name     string
		request  PaginationRequest
		expected int
	}{
		{
			name: "page 5",
			request: PaginationRequest{
				Page: 5,
			},
			expected: 5,
		},
		{
			name:     "zero value",
			request:  PaginationRequest{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				actual := tt.request.GetPage()
				assert.Equal(t, tt.expected, actual)
			},
		)
	}
}

func TestPaginationRequest_Default(t *testing.T) {
	tests := []struct {
		name            string
		input           PaginationRequest
		expectedPage    int
		expectedPerPage int
	}{
		{
			name:            "both zero values",
			input:           PaginationRequest{},
			expectedPage:    1,
			expectedPerPage: 10,
		},
		{
			name: "page zero",
			input: PaginationRequest{
				PerPage: 25,
			},
			expectedPage:    1,
			expectedPerPage: 25,
		},
		{
			name: "per page zero",
			input: PaginationRequest{
				Page: 3,
			},
			expectedPage:    3,
			expectedPerPage: 10,
		},
		{
			name: "no zero values",
			input: PaginationRequest{
				Page:    2,
				PerPage: 15,
			},
			expectedPage:    2,
			expectedPerPage: 15,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.input.Default()
				assert.Equal(t, tt.expectedPage, tt.input.Page)
				assert.Equal(t, tt.expectedPerPage, tt.input.PerPage)
			},
		)
	}
}

func TestPaginationResponse_Fields(t *testing.T) {
	resp := PaginationResponse{
		Page:    2,
		PerPage: 20,
		MaxPage: 5,
		Count:   100,
	}

	assert.Equal(t, 2, resp.Page)
	assert.Equal(t, 20, resp.PerPage)
	assert.Equal(t, int64(5), resp.MaxPage)
	assert.Equal(t, int64(100), resp.Count)
}
