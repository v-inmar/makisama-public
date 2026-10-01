package image_utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetExtFromContentType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "jpeg", input: "image/jpeg", expected: ".jpg"},
		{name: "png", input: "image/png", expected: ".png"},
		{name: "webp", input: "image/webp", expected: ".webp"},
		{name: "unsupported", input: "image/gif", expected: ""},
		{name: "empty", input: "", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExtFromContentType(tt.input); got != tt.expected {
				t.Fatalf("GetExtFromContentType(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDeleteImageFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "sample-image.png")

	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if err := DeleteImageFile(filePath); err != nil {
		t.Fatalf("DeleteImageFile returned unexpected error: %v", err)
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("expected file to be removed, got err=%v", err)
	}
}

func TestDeleteImageFileMissingFile(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "does-not-exist.png")

	if err := DeleteImageFile(missingPath); !os.IsNotExist(err) {
		t.Fatalf("DeleteImageFile(%q) error = %v, want file not exist", missingPath, err)
	}
}
