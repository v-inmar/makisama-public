package slice_utils

import (
	"reflect"
	"testing"
)

func TestAppendToSlice(t *testing.T) {
	tests := []struct {
		name      string
		current   []string
		newString string
		expected  []string
	}{
		{
			name:      "append to existing slice",
			current:   []string{"alpha", "beta"},
			newString: "gamma",
			expected:  []string{"alpha", "beta", "gamma"},
		},
		{
			name:      "append to nil slice",
			current:   nil,
			newString: "first",
			expected:  []string{"first"},
		},
		{
			name:      "append empty string",
			current:   []string{"alpha"},
			newString: "",
			expected:  []string{"alpha", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := AppendToSlice(tt.current, tt.newString)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Fatalf("AppendToSlice(%v, %q) = %v, want %v", tt.current, tt.newString, actual, tt.expected)
			}
		})
	}
}
