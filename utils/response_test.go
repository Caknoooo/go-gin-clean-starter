package utils

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildResponseSuccess(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		data     any
		expected Response
	}{
		{
			name:    "Success with data",
			message: "Operation successful",
			data:    map[string]string{"key": "value"},
			expected: Response{
				Status:  true,
				Message: "Operation successful",
				Data:    map[string]string{"key": "value"},
			},
		},
		{
			name:    "Success with empty data",
			message: "Operation successful",
			data:    nil,
			expected: Response{
				Status:  true,
				Message: "Operation successful",
				Data:    nil,
			},
		},
		{
			name:    "Success with empty message",
			message: "",
			data:    "some data",
			expected: Response{
				Status:  true,
				Message: "",
				Data:    "some data",
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := BuildResponseSuccess(tt.message, tt.data)
				assert.Equal(t, tt.expected, result)
			},
		)
	}
}

func TestBuildResponseFailed(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		err      string
		data     any
		expected Response
	}{
		{
			name:    "Failed with error string",
			message: "Operation failed",
			err:     "something went wrong",
			data:    nil,
			expected: Response{
				Status:  false,
				Message: "Operation failed",
				Error:   "something went wrong",
				Data:    nil,
			},
		},
		{
			name:    "Failed with error object",
			message: "Operation failed",
			err:     errors.New("something went wrong").Error(),
			data:    map[string]string{"key": "value"},
			expected: Response{
				Status:  false,
				Message: "Operation failed",
				Error:   "something went wrong",
				Data:    map[string]string{"key": "value"},
			},
		},
		{
			name:    "Failed with empty error",
			message: "Operation failed",
			err:     "",
			data:    "some data",
			expected: Response{
				Status:  false,
				Message: "Operation failed",
				Error:   "",
				Data:    "some data",
			},
		},
		{
			name:    "Failed with empty message",
			message: "",
			err:     "error occurred",
			data:    nil,
			expected: Response{
				Status:  false,
				Message: "",
				Error:   "error occurred",
				Data:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := BuildResponseFailed(tt.message, tt.err, tt.data)
				assert.Equal(t, tt.expected, result)
			},
		)
	}
}

func TestEmptyObj(t *testing.T) {
	// Test that EmptyObj is indeed an empty struct by checking its type and size
	var empty EmptyObj

	// Check the type is exactly EmptyObj
	assert.Equal(t, "EmptyObj", reflect.TypeOf(empty).Name())

	// Check the size is 0 (empty struct)
	assert.Equal(t, 0, int(reflect.TypeOf(empty).Size()))

	// Compare with another EmptyObj instance
	assert.True(t, reflect.DeepEqual(empty, EmptyObj{}))
}
