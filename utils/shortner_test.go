package utils

import (
	"testing"
)

func TestGenerateShortCode(t *testing.T) {
	code := GenerateShortCode()

	if len(code) != 6 {
		t.Errorf("Expected code length 6, got %d", len(code))
	}

	if code == "" {
		t.Error("Expected non empty code")
	}
}
