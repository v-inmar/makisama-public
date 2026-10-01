package number_utils

import "testing"

func TestMakeSliceFromInt(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected []int
	}{
		{name: "zero length", input: 0, expected: []int{}},
		{name: "positive length", input: 5, expected: []int{1, 2, 3, 4, 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input < 0 {
				defer func() {
					if r := recover(); r == nil {
						t.Fatalf("MakeSliceFromInt(%d) should panic for negative input", tt.input)
					}
				}()
			}

			actual := MakeSliceFromInt(tt.input)
			if tt.input < 0 {
				return
			}
			if len(actual) != len(tt.expected) {
				t.Fatalf("MakeSliceFromInt(%d) length = %d, want %d", tt.input, len(actual), len(tt.expected))
			}
			for i := range actual {
				if actual[i] != tt.expected[i] {
					t.Fatalf("MakeSliceFromInt(%d)[%d] = %d, want %d", tt.input, i, actual[i], tt.expected[i])
				}
			}
		})
	}
}

func TestAddition(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{name: "positive", input: 5, expected: 6},
		{name: "zero", input: 0, expected: 1},
		{name: "negative", input: -2, expected: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Addition(tt.input)
			if actual != tt.expected {
				t.Fatalf("Addition(%d) = %d, want %d", tt.input, actual, tt.expected)
			}
		})
	}
}

func TestSubtraction(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{name: "positive", input: 5, expected: 4},
		{name: "zero", input: 0, expected: -1},
		{name: "negative", input: -2, expected: -3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Subtraction(tt.input)
			if actual != tt.expected {
				t.Fatalf("Subtraction(%d) = %d, want %d", tt.input, actual, tt.expected)
			}
		})
	}
}
