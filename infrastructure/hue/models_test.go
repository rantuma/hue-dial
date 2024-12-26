package hue_test

import (
	"testing"

	"github.com/rantuma/hue-dial/infrastructure/hue"

	"github.com/stretchr/testify/assert"
)

func TestErrors_String(t *testing.T) {
	tests := []struct {
		name     string
		errors   hue.Errors
		expected string
	}{
		{
			name:     "No errors",
			errors:   hue.Errors{},
			expected: "No errors",
		},
		{
			name: "Single error",
			errors: hue.Errors{
				{Description: "Invalid input"},
			},
			expected: "Invalid input",
		},
		{
			name: "Multiple errors",
			errors: hue.Errors{
				{Description: "Invalid input"},
				{Description: "Missing field name"},
				{Description: "Unknown error occurred"},
			},
			expected: "Invalid input; Missing field name; Unknown error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.errors.String())
		})
	}
}
