package randoms_utils

import (
	"strings"
	"testing"
)

func TestGenRandNumBetween(t *testing.T) {
	for i := 0; i < 100; i++ {
		value := GenRandNumBetween(10, 20)
		if value < 10 || value > 20 {
			t.Fatalf("GenRandNumBetween(10, 20) returned %d, want a value between 10 and 20", value)
		}
	}
}

func TestSecureRandString(t *testing.T) {
	value, err := SecureRandString(5, 8)
	if err != nil {
		t.Fatalf("SecureRandString returned unexpected error: %v", err)
	}

	if len(value) < 5 || len(value) > 8 {
		t.Fatalf("SecureRandString returned length %d, want between 5 and 8", len(value))
	}

	for _, ch := range value {
		if !strings.ContainsRune(charset, ch) {
			t.Fatalf("SecureRandString produced invalid character %q", ch)
		}
	}
}

func TestSecureRandStringWithPool(t *testing.T) {
	pool := "abc123"
	value, err := SecureRandStringWithPool(pool, 3, 5)
	if err != nil {
		t.Fatalf("SecureRandStringWithPool returned unexpected error: %v", err)
	}

	if len(value) < 3 || len(value) > 5 {
		t.Fatalf("SecureRandStringWithPool returned length %d, want between 3 and 5", len(value))
	}

	for _, ch := range value {
		if !strings.ContainsRune(pool, ch) {
			t.Fatalf("SecureRandStringWithPool produced invalid character %q", ch)
		}
	}
}

func TestGenerateCryptoRandValueHelpers(t *testing.T) {
	tests := []struct {
		name      string
		generator func() (string, error)
		minLength int
	}{
		{name: "reset token", generator: GenerateResetToken, minLength: 32},
		{name: "authid value", generator: GenerateAuthidValue, minLength: 32},
		{name: "pid value", generator: GeneratePidValue, minLength: 12},
		{name: "feedback slug", generator: GenerateFeedbackSlugValue, minLength: 64},
		{name: "image filename", generator: GenerateImageFilename, minLength: 32},
		{name: "confirm token", generator: GenerateConfirmToken, minLength: 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.generator()
			if err != nil {
				t.Fatalf("%s returned unexpected error: %v", tt.name, err)
			}
			if len(value) < tt.minLength {
				t.Fatalf("%s returned length %d, want at least %d", tt.name, len(value), tt.minLength)
			}
		})
	}
}
